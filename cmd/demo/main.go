package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/holden/vga-go/internal/effects"
	"github.com/holden/vga-go/internal/music"
	demosync "github.com/holden/vga-go/internal/sync"
	"github.com/holden/vga-go/internal/vga"

	"github.com/ebitengine/oto/v3"
)

var version = "dev"

const Scale = 3

type Demo struct {
	fb           *vga.Framebuffer
	screen       *ebiten.Image
	sequencer    *demosync.Sequencer
	player       *music.Player
	otoCtx       *oto.Context
	otoPlayer    *oto.Player
	lastTime     time.Time
	showDebug    bool
	quit         chan struct{} // closed by signal handler
	shuttingDown bool
	fadeStart    time.Time
}

var debugMode bool

func NewDemo(modFile, cueFile string) (*Demo, error) {
	fb := vga.NewFramebuffer(vga.DefaultPalette())

	// Create effects
	plasma := effects.NewPlasma()
	fire := effects.NewFire()
	tunnel := effects.NewTunnel()
	starfield := effects.NewStarfield()
	sineScroller := effects.NewSineScroller("HELLO DEMOSCENE! THIS IS VGA-GO - A DEMO ENGINE IN GO!    ")
	bigScroller := effects.NewBigScroller("VGA-GO DEMO ENGINE    ")
	efx := []effects.Effect{plasma, fire, tunnel, starfield, sineScroller, bigScroller}

	var timeline *demosync.Timeline
	if cueFile != "" {
		tl, err := demosync.LoadCueFile(cueFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load cue file: %w", err)
		}
		timeline = tl
	} else {
		// Default timeline
		timeline = demosync.NewTimeline([]demosync.Cue{
			{Pos: demosync.Position{Order: 0, Row: 0}, EffectIdx: 0, Transition: "cut"},
			{Pos: demosync.Position{Order: 1, Row: 0}, EffectIdx: 3, Transition: "cut"},
			{Pos: demosync.Position{Order: 2, Row: 0}, EffectIdx: 2, Transition: "cut"},
			{Pos: demosync.Position{Order: 3, Row: 0}, EffectIdx: 1, Transition: "cut"},
			{Pos: demosync.Position{Order: 4, Row: 0}, EffectIdx: 0, Transition: "fade", FadeDur: 2.0},
		})
	}

	seq := demosync.NewSequencer(efx, timeline)
	seq.InitFirst(fb)

	d := &Demo{
		fb:        fb,
		screen:    ebiten.NewImage(vga.Width, vga.Height),
		sequencer: seq,
		lastTime:  time.Now(),
		quit:      make(chan struct{}),
	}

	// Load and start music if a mod file is provided
	if modFile != "" {
		player, err := music.NewPlayer(modFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load module %s: %w", modFile, err)
		}
		d.player = player
		player.SetDebug(debugMode)

		if debugMode {
			log.Printf("[main] loaded module: %s", modFile)
		}

		// Setup audio output via oto
		op := &oto.NewContextOptions{
			SampleRate:   music.SampleRate,
			ChannelCount: 2,
			Format:       oto.FormatSignedInt16LE,
		}
		otoCtx, readyChan, err := oto.NewContext(op)
		if err != nil {
			player.Stop()
			return nil, fmt.Errorf("failed to create audio context: %w", err)
		}
		<-readyChan
		if debugMode {
			log.Printf("[main] oto audio context ready, sample rate=%d", music.SampleRate)
		}
		d.otoCtx = otoCtx
		d.otoPlayer = otoCtx.NewPlayer(player)
		d.otoPlayer.SetBufferSize(music.SampleRate * 4 * 2) // 2 seconds of buffer
		d.otoPlayer.Play()

		if debugMode {
			log.Printf("[main] oto player started, starting render loop")
		}

		// Start the music render loop
		player.Start()
	}

	return d, nil
}

func (d *Demo) Update() error {
	// Check for signal-triggered quit
	select {
	case <-d.quit:
		d.startShutdown()
	default:
	}

	if !d.shuttingDown && ebiten.IsKeyPressed(ebiten.KeyEscape) {
		d.startShutdown()
	}

	if inputJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inputJustPressed(ebiten.KeyF1) {
		d.showDebug = !d.showDebug
	}

	if d.shuttingDown {
		d.updateFade()
		if time.Since(d.fadeStart) >= 5*time.Second {
			return ebiten.Termination
		}
		return nil
	}

	now := time.Now()
	dt := now.Sub(d.lastTime).Seconds()
	d.lastTime = now

	// Get sync state from music player (or empty if no music)
	var syncState music.FrameInfo
	if d.player != nil {
		syncState = d.player.SyncState()
	}

	d.sequencer.Update(dt, syncState, d.fb)

	return nil
}

func (d *Demo) startShutdown() {
	if d.shuttingDown {
		return
	}
	d.shuttingDown = true
	d.fadeStart = time.Now()
	log.Printf("[main] starting shutdown, fading over 5s...")
}

func (d *Demo) updateFade() {
	elapsed := time.Since(d.fadeStart)
	fadeDur := 5 * time.Second
	progress := 1.0 - float64(elapsed)/float64(fadeDur)
	if progress < 0 {
		progress = 0
	}

	if d.otoPlayer != nil {
		d.otoPlayer.SetVolume(progress)
	}

	for i := range d.fb.Palette {
		c := d.fb.Palette[i]
		d.fb.Palette[i] = color.RGBA{
			R: uint8(float64(c.R) * progress),
			G: uint8(float64(c.G) * progress),
			B: uint8(float64(c.B) * progress),
			A: c.A,
		}
	}

	if elapsed >= fadeDur {
		log.Printf("[main] fade complete")
		return
	}

	if int(elapsed.Seconds()) != int((elapsed - 16*time.Millisecond).Seconds()) {
		log.Printf("[main] fading... %.0f%%", progress*100)
	}
}

func (d *Demo) Draw(screen *ebiten.Image) {
	d.sequencer.Draw(d.fb)
	d.screen.WritePixels(d.fb.RGBA())
	screen.DrawImage(d.screen, nil)

	if d.showDebug && d.player != nil {
		info := d.player.SyncState()
		ebitenutil.DebugPrint(screen, fmt.Sprintf(
			"Ord:%02d Pat:%02d Row:%02d/%02d BPM:%d Spd:%d\nFPS:%.0f",
			info.Order, info.Pattern, info.Row, info.NumRows,
			info.BPM, info.Speed, ebiten.ActualFPS(),
		))
	} else if d.showDebug {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS:%.0f (no music)", ebiten.ActualFPS()))
	}
}

func (d *Demo) Layout(outsideWidth, outsideHeight int) (int, int) {
	return vga.Width, vga.Height
}

func (d *Demo) Close() {
	log.Printf("[main] shutdown: stopping player...")
	if d.player != nil {
		d.player.Stop()
	}
	log.Printf("[main] shutdown complete")
}

var keyStates = map[ebiten.Key]bool{}

func inputJustPressed(key ebiten.Key) bool {
	pressed := ebiten.IsKeyPressed(key)
	was := keyStates[key]
	keyStates[key] = pressed
	return pressed && !was
}

func main() {
	modFile := flag.String("mod", "", "Path to MOD/S3M/XM/IT tracker module file")
	cueFile := flag.String("cue", "", "Path to JSON cue file (demo timeline)")
	fullscreen := flag.Bool("fullscreen", false, "Start in fullscreen mode")
	debug := flag.Bool("debug", false, "Enable debug logging for music playback")
	flag.Parse()

	debugMode = *debug

	log.Printf("VGA-GO Demo Engine %s", version)

	demo, err := NewDemo(*modFile, *cueFile)
	if err != nil {
		log.Fatal(err)
	}
	defer demo.Close()

	// Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("[main] received signal: %s", sig)
		close(demo.quit)
	}()

	ebiten.SetWindowSize(vga.Width*Scale, vga.Height*Scale)
	ebiten.SetWindowTitle("VGA-GO Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(*fullscreen)

	if err := ebiten.RunGame(demo); err != nil {
		log.Printf("[main] game loop ended: %v", err)
	}

	log.Printf("[main] calling close...")
	demo.Close()
	log.Printf("[main] exiting")
}
