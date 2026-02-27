package effects

import (
	"math"

	"github.com/holden/vga-go/internal/music"
	"github.com/holden/vga-go/internal/vga"
)

// Tunnel is a classic texture-mapped tunnel/wormhole effect.
type Tunnel struct {
	// Pre-computed lookup tables
	angleLUT [vga.Size]float64
	depthLUT [vga.Size]float64
	time     float64
}

func NewTunnel() *Tunnel {
	t := &Tunnel{}
	cx := float64(vga.Width) / 2.0
	cy := float64(vga.Height) / 2.0

	for y := 0; y < vga.Height; y++ {
		for x := 0; x < vga.Width; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)
			idx := y*vga.Width + x

			if dist < 1.0 {
				dist = 1.0
			}
			t.depthLUT[idx] = 128.0 / dist
			t.angleLUT[idx] = math.Atan2(dy, dx) / math.Pi * 128.0
		}
	}
	return t
}

func (t *Tunnel) Init(fb *vga.Framebuffer) {
	fb.SetPalette(vga.PlasmaPalette())
}

func (t *Tunnel) Update(dt float64, sync music.FrameInfo) {
	speed := 1.0
	if sync.BPM > 0 {
		speed = float64(sync.BPM) / 120.0
	}
	t.time += dt * speed
}

func (t *Tunnel) Draw(fb *vga.Framebuffer) {
	shiftU := t.time * 50.0
	shiftV := t.time * 30.0

	for i := 0; i < vga.Size; i++ {
		u := int(t.depthLUT[i]+shiftU) & 255
		v := int(t.angleLUT[i]+shiftV) & 255

		// XOR texture — classic demoscene pattern
		fb.Pixels[i] = byte(u ^ v)
	}
}
