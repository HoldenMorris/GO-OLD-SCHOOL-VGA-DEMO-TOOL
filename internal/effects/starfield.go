package effects

import (
	"math/rand"

	"github.com/holden/vga-go/internal/music"
	"github.com/holden/vga-go/internal/vga"
)

const numStars = 512

type star struct {
	x, y, z float64
}

// Starfield is a classic 3D parallax starfield flying through space.
type Starfield struct {
	stars [numStars]star
	speed float64
}

func NewStarfield() *Starfield {
	sf := &Starfield{speed: 1.0}
	for i := range sf.stars {
		sf.stars[i] = sf.randomStar()
		sf.stars[i].z = rand.Float64() * 256.0 // spread initial depth
	}
	return sf
}

func (sf *Starfield) randomStar() star {
	return star{
		x: (rand.Float64() - 0.5) * 640.0,
		y: (rand.Float64() - 0.5) * 400.0,
		z: 256.0,
	}
}

func (sf *Starfield) Init(fb *vga.Framebuffer) {
	fb.SetPalette(vga.DefaultPalette())
}

func (sf *Starfield) Update(dt float64, sync music.FrameInfo) {
	sf.speed = 200.0
	if sync.BPM > 0 {
		sf.speed = float64(sync.BPM) * 1.5
		// Warp speed on beat
		if sync.Speed > 0 && sync.Frame == 0 {
			sf.speed *= 3.0
		}
	}

	// Move stars toward the viewer
	for i := range sf.stars {
		sf.stars[i].z -= sf.speed * dt
		if sf.stars[i].z <= 1.0 {
			sf.stars[i] = sf.randomStar()
		}
	}
}

func (sf *Starfield) Draw(fb *vga.Framebuffer) {
	fb.Clear(0) // black background

	cx := float64(vga.Width) / 2.0
	cy := float64(vga.Height) / 2.0

	for i := range sf.stars {
		s := &sf.stars[i]
		// Project 3D → 2D
		sx := int(s.x/s.z*128.0 + cx)
		sy := int(s.y/s.z*128.0 + cy)

		if sx < 0 || sx >= vga.Width || sy < 0 || sy >= vga.Height {
			continue
		}

		// Brightness based on depth (closer = brighter)
		bright := int(255.0 - s.z)
		if bright < 0 {
			bright = 0
		}
		// Map to white/gray palette entries (7=light gray, 8=dark gray, 15=white)
		var colorIdx byte
		switch {
		case bright > 200:
			colorIdx = 15 // white
		case bright > 128:
			colorIdx = 7 // light gray
		default:
			colorIdx = 8 // dark gray
		}

		fb.Pixels[sy*vga.Width+sx] = colorIdx
	}
}
