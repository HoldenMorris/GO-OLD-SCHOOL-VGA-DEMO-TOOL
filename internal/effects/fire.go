package effects

import (
	"math/rand"

	"github.com/holden/vga-go/internal/music"
	"github.com/holden/vga-go/internal/vga"
)

// Fire is a classic bottom-up heat propagation fire effect.
type Fire struct {
	heat     [vga.Size]int
	intensity float64
}

func NewFire() *Fire {
	return &Fire{intensity: 1.0}
}

func (f *Fire) Init(fb *vga.Framebuffer) {
	fb.SetPalette(vga.FirePalette())
	for i := range f.heat {
		f.heat[i] = 0
	}
}

func (f *Fire) Update(dt float64, sync music.FrameInfo) {
	f.intensity = 1.0
	if sync.BPM > 0 {
		// Boost intensity on beats
		if sync.Speed > 0 && sync.Frame == 0 {
			f.intensity = 2.0
		}
		// Use channel volumes for extra flicker
		maxVol := 0
		for i := 0; i < sync.NumChannels; i++ {
			if sync.ChannelVol[i] > maxVol {
				maxVol = sync.ChannelVol[i]
			}
		}
		f.intensity += float64(maxVol) / 255.0
	}
}

func (f *Fire) Draw(fb *vga.Framebuffer) {
	// Seed the bottom row with random hot pixels
	base := int(200.0 * f.intensity)
	if base > 255 {
		base = 255
	}
	for x := 0; x < vga.Width; x++ {
		f.heat[(vga.Height-1)*vga.Width+x] = rand.Intn(base + 1)
		if vga.Height >= 2 {
			f.heat[(vga.Height-2)*vga.Width+x] = rand.Intn(base + 1)
		}
	}

	// Propagate heat upward with averaging + decay
	for y := 0; y < vga.Height-2; y++ {
		for x := 0; x < vga.Width; x++ {
			left := x - 1
			if left < 0 {
				left = 0
			}
			right := x + 1
			if right >= vga.Width {
				right = vga.Width - 1
			}

			sum := f.heat[(y+1)*vga.Width+left] +
				f.heat[(y+1)*vga.Width+x] +
				f.heat[(y+1)*vga.Width+right] +
				f.heat[(y+2)*vga.Width+x]

			avg := sum / 4
			decay := avg - 2
			if decay < 0 {
				decay = 0
			}
			f.heat[y*vga.Width+x] = decay
		}
	}

	// Write to framebuffer
	for i := range f.heat {
		v := f.heat[i]
		if v > 255 {
			v = 255
		}
		fb.Pixels[i] = byte(v)
	}
}
