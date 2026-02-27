package effects

import (
	"math"

	"github.com/holden/vga-go/internal/music"
	"github.com/holden/vga-go/internal/vga"
)

type SineScroller struct {
	text      string
	offset    float64
	time      float64
	charWidth int
	speed     float64
}

func NewSineScroller(text string) *SineScroller {
	return &SineScroller{
		text:      text + "  ",
		offset:    float64(vga.Width),
		charWidth: 8,
		speed:     60,
	}
}

func (s *SineScroller) Init(fb *vga.Framebuffer) {
	fb.SetPalette(vga.DefaultPalette())
	vga.LoadFontFromPNG("assets/font1.png")
}

func (s *SineScroller) Update(dt float64, sync music.FrameInfo) {
	s.time += dt
	speed := s.speed
	if sync.BPM > 0 {
		speed = float64(sync.BPM) * 0.8
	}
	s.offset -= dt * speed
	if s.offset < -float64(len(s.text)*s.charWidth) {
		s.offset = float64(vga.Width)
	}
}

func (s *SineScroller) Draw(fb *vga.Framebuffer) {
	centerY := vga.Height / 2
	amp := 20.0 + math.Sin(s.time*2)*10
	freq := 0.1 + math.Sin(s.time)*0.05

	for i, ch := range s.text {
		charX := int(s.offset) + i*s.charWidth
		for fx := 0; fx < 8; fx++ {
			for fy := 0; fy < 8; fy++ {
				font := vga.CP437Font[byte(ch)]
				if font[fy]&(1<<uint(fx)) != 0 {
					sineY := math.Sin(float64(charX+fx)*freq+s.time*3) * amp
					py := centerY - 4 + fy + int(sineY)
					px := charX + fx
					if px >= 0 && px < vga.Width && py >= 0 && py < vga.Height {
						fb.SetPixel(px, py, 255)
					}
				}
			}
		}
	}
}

type BigScroller struct {
	text   string
	offset float64
	time   float64
	speed  float64
	scale  int
}

func NewBigScroller(text string) *BigScroller {
	return &BigScroller{
		text:   text + "  ",
		offset: float64(vga.Width),
		speed:  60,
		scale:  3,
	}
}

func (b *BigScroller) Init(fb *vga.Framebuffer) {
	fb.SetPalette(vga.DefaultPalette())
	vga.LoadFontFromPNG("assets/font1.png")
}

func (b *BigScroller) Update(dt float64, sync music.FrameInfo) {
	b.time += dt
	speed := b.speed
	if sync.BPM > 0 {
		speed = float64(sync.BPM) * 0.8
	}
	b.offset -= dt * speed
	if b.offset < -float64(len(b.text)*8*b.scale) {
		b.offset = float64(vga.Width)
	}
}

func (b *BigScroller) Draw(fb *vga.Framebuffer) {
	centerY := (vga.Height - 8*b.scale) / 2

	for i, ch := range b.text {
		charX := int(b.offset) + i*8*b.scale
		charSprite := vga.CharToSprite(byte(ch))
		fb.DrawSpriteScaled(charX, centerY, charSprite, b.scale)
	}
}
