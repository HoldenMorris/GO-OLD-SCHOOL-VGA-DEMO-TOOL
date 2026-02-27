package music

import (
	"io"
	"log"
	"sync"
	"time"
)

const (
	SampleRate = 44100
	// bufSize needs to hold several seconds of audio to handle timing jitter.
	// 44100 Hz * 2 channels * 2 bytes * 4 seconds = 705600 bytes
	bufSize = 1 << 20 // 1MB ring buffer
)

// Player manages tracker module playback and exposes sync state.
type Player struct {
	ctx  *Context
	mu   sync.RWMutex
	info FrameInfo

	ring     []byte
	ringR    int
	ringW    int
	ringLen  int
	ringMu   sync.Mutex
	ringCond *sync.Cond

	playing   bool
	done      chan struct{}
	stopOnce  sync.Once
	debug     bool

	volume    float64
	volumeMut sync.Mutex
	fadeMut   sync.Mutex

	framesRendered uint64
	bytesWritten   uint64
	bytesRead      uint64
}

// NewPlayer creates a player for the given module file path.
func NewPlayer(path string) (*Player, error) {
	ctx := NewContext()
	if err := ctx.LoadModule(path); err != nil {
		ctx.Close()
		return nil, err
	}
	if err := ctx.StartPlayer(SampleRate); err != nil {
		ctx.ReleaseModule()
		ctx.Close()
		return nil, err
	}

	p := &Player{
		ctx:    ctx,
		ring:   make([]byte, bufSize),
		done:   make(chan struct{}),
		volume: 1.0,
	}
	p.ringCond = sync.NewCond(&p.ringMu)
	return p, nil
}

// NewPlayerFromMemory creates a player from module data in memory.
func NewPlayerFromMemory(data []byte) (*Player, error) {
	ctx := NewContext()
	if err := ctx.LoadModuleFromMemory(data); err != nil {
		ctx.Close()
		return nil, err
	}
	if err := ctx.StartPlayer(SampleRate); err != nil {
		ctx.ReleaseModule()
		ctx.Close()
		return nil, err
	}

	p := &Player{
		ctx:    ctx,
		ring:   make([]byte, bufSize),
		done:   make(chan struct{}),
		volume: 1.0,
	}
	p.ringCond = sync.NewCond(&p.ringMu)
	return p, nil
}

// SetDebug enables verbose logging of playback state.
func (p *Player) SetDebug(on bool) {
	p.debug = on
}

// Start begins rendering audio frames in a background goroutine.
func (p *Player) Start() {
	p.playing = true
	go p.renderLoop()
	if p.debug {
		go p.debugLoop()
	}
}

// renderLoop continuously renders audio frames from libxmp into the ring buffer.
func (p *Player) renderLoop() {
	defer close(p.done)

	if p.debug {
		log.Printf("[music] render loop started, sample rate=%d, ring buffer=%d bytes", SampleRate, bufSize)
	}

	for p.playing {
		if !p.ctx.PlayFrame() {
			if p.debug {
				log.Printf("[music] PlayFrame returned false (end of module), frames rendered: %d", p.framesRendered)
			}
			return
		}
		p.framesRendered++

		// Update sync info
		info := p.ctx.GetFrameInfo()
		p.mu.Lock()
		p.info = info
		p.mu.Unlock()

		if p.debug && p.framesRendered <= 5 {
			log.Printf("[music] frame %d: ord=%d pat=%d row=%d/%d bpm=%d spd=%d time=%dms loop=%d",
				p.framesRendered, info.Order, info.Pattern, info.Row, info.NumRows,
				info.BPM, info.Speed, info.TimeMs, info.LoopCount)
		}

		// Get PCM buffer from libxmp
		buf := p.ctx.GetBuffer()
		if len(buf) == 0 {
			if p.debug {
				log.Printf("[music] frame %d: empty buffer from libxmp", p.framesRendered)
			}
			continue
		}

		// Write to ring buffer, waiting via cond var if full
		p.ringMu.Lock()
		written := 0
		for written < len(buf) {
			// Wait for space if ring is full
			for p.ringLen >= bufSize && p.playing {
				p.ringCond.Wait() // woken by Read() consuming data or Stop() broadcasting
			}
			if !p.playing {
				p.ringMu.Unlock()
				if p.debug {
					log.Printf("[music] render loop interrupted during write, frames=%d", p.framesRendered)
				}
				return
			}

			// Write as much as we can
			space := bufSize - p.ringLen
			toWrite := len(buf) - written
			if toWrite > space {
				toWrite = space
			}

			for i := 0; i < toWrite; i++ {
				p.ring[p.ringW] = buf[written+i]
				p.ringW = (p.ringW + 1) % bufSize
			}
			p.ringLen += toWrite
			p.bytesWritten += uint64(toWrite)
			written += toWrite
		}
		p.ringMu.Unlock()
	}

	if p.debug {
		log.Printf("[music] render loop exiting, frames rendered: %d, bytes written: %d",
			p.framesRendered, p.bytesWritten)
	}
}

// debugLoop periodically logs playback stats.
func (p *Player) debugLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			p.ringMu.Lock()
			ringPct := float64(p.ringLen) / float64(bufSize) * 100.0
			p.ringMu.Unlock()

			info := p.SyncState()
			log.Printf("[music] ord=%d pat=%d row=%02d/%02d bpm=%d | ring: %.1f%% full | rendered=%d written=%dKB read=%dKB",
				info.Order, info.Pattern, info.Row, info.NumRows, info.BPM,
				ringPct, p.framesRendered,
				p.bytesWritten/1024, p.bytesRead/1024)
		case <-p.done:
			return
		}
	}
}

// Read implements io.Reader for oto.Player consumption.
// Returns 16-bit signed stereo interleaved PCM at 44100 Hz.
func (p *Player) Read(buf []byte) (int, error) {
	p.ringMu.Lock()
	n := 0
	for n < len(buf) && p.ringLen > 0 {
		buf[n] = p.ring[p.ringR]
		p.ringR = (p.ringR + 1) % bufSize
		p.ringLen--
		n++
	}
	p.bytesRead += uint64(n)
	if n > 0 {
		p.ringCond.Signal() // wake renderLoop if it's waiting for space
	}
	p.ringMu.Unlock()

	if n == 0 {
		select {
		case <-p.done:
			if p.debug {
				log.Printf("[music] Read: EOF, total read=%dKB", p.bytesRead/1024)
			}
			return 0, io.EOF
		default:
			// No data yet but still playing — return silence
			for i := range buf {
				buf[i] = 0
			}
			return len(buf), nil
		}
	}

	// If we couldn't fill the entire buffer, pad with silence
	// This prevents oto from getting partial frames
	if n < len(buf) {
		for i := n; i < len(buf); i++ {
			buf[i] = 0
		}
		return len(buf), nil
	}

	return n, nil
}

// SyncState returns the current playback sync info (thread-safe).
func (p *Player) SyncState() FrameInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info
}

// Stop stops playback and releases resources. Safe to call multiple times.
func (p *Player) Stop() {
	p.stopOnce.Do(func() {
		log.Printf("[music] stopping playback...")
		p.playing = false

		// Wake renderLoop if it's blocked waiting for ring buffer space
		p.ringCond.Broadcast()

		// Wait for renderLoop to exit, with timeout
		select {
		case <-p.done:
			if p.debug {
				log.Printf("[music] render loop exited cleanly")
			}
		case <-time.After(2 * time.Second):
			log.Printf("[music] warning: render loop did not exit within 2s")
		}

		p.ctx.EndPlayer()
		p.ctx.ReleaseModule()
		p.ctx.Close()
		log.Printf("[music] cleanup complete (rendered %d frames, %dKB audio)",
			p.framesRendered, p.bytesWritten/1024)
	})
}

// Volume returns the current volume (0.0 to 1.0).
func (p *Player) Volume() float64 {
	p.volumeMut.Lock()
	defer p.volumeMut.Unlock()
	return p.volume
}

// SetVolume sets the playback volume (0.0 to 1.0).
func (p *Player) SetVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.volumeMut.Lock()
	p.volume = v
	p.volumeMut.Unlock()
}

// FadeOut fades volume from current level to 0 over duration seconds.
// Returns a channel that closes when fade is complete.
func (p *Player) FadeOut(duration time.Duration) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		startVol := p.Volume()
		if startVol <= 0 {
			close(done)
			return
		}
		startTime := time.Now()
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		for {
			elapsed := time.Since(startTime)
			if elapsed >= duration {
				p.SetVolume(0)
				close(done)
				return
			}
			t := 1.0 - float64(elapsed)/float64(duration)
			p.SetVolume(startVol * t)
			<-ticker.C
		}
	}()
	return done
}

// Done returns a channel that closes when playback finishes.
func (p *Player) Done() <-chan struct{} {
	return p.done
}
