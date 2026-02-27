package effects

import (
	"math"

	"github.com/holden/vga-go/internal/music"
	"github.com/holden/vga-go/internal/vga"
)

// Plasma is a classic demoscene sine-based plasma effect.
type Plasma struct {
	time     float64
	sinTable [256]float64
}

func NewPlasma() *Plasma {
	p := &Plasma{}
	for i := range p.sinTable {
		p.sinTable[i] = math.Sin(float64(i) * math.Pi * 2.0 / 256.0)
	}
	return p
}

func (p *Plasma) Init(fb *vga.Framebuffer) {
	fb.SetPalette(vga.PlasmaPalette())
}

func (p *Plasma) Update(dt float64, sync music.FrameInfo) {
	speed := 1.0
	if sync.BPM > 0 {
		speed = float64(sync.BPM) / 120.0
	}
	// Pulse on beat
	if sync.Speed > 0 && sync.Frame == 0 {
		speed *= 1.5
	}
	p.time += dt * speed
}

func (p *Plasma) Draw(fb *vga.Framebuffer) {
	t := p.time * 50.0

	for y := 0; y < vga.Height; y++ {
		fy := float64(y)
		off := y * vga.Width
		for x := 0; x < vga.Width; x++ {
			fx := float64(x)

			v := p.sin(fx*0.08 + t*0.03)
			v += p.sin(fy*0.07 + t*0.02)
			v += p.sin((fx+fy)*0.05 + t*0.025)
			v += p.sin(math.Sqrt(fx*fx+fy*fy)*0.06 + t*0.015)

			// Map [-4, 4] → [0, 255]
			idx := byte((v + 4.0) * 31.875)
			fb.Pixels[off+x] = idx
		}
	}
}

func (p *Plasma) sin(x float64) float64 {
	// Use the lookup table with linear interpolation
	x = math.Mod(x, 2*math.Pi)
	if x < 0 {
		x += 2 * math.Pi
	}
	pos := x * 256.0 / (2 * math.Pi)
	i := int(pos) & 255
	frac := pos - math.Floor(pos)
	next := (i + 1) & 255
	return p.sinTable[i]*(1-frac) + p.sinTable[next]*frac
}
