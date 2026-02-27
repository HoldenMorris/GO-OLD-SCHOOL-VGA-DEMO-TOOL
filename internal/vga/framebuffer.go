package vga

import (
	"image/color"
	"time"
)

const (
	Width  = 320
	Height = 200
	Size   = Width * Height
)

// Framebuffer is a 320x200 indexed-color buffer, emulating VGA Mode 13h.
// Each byte is a palette index (0-255).
type Framebuffer struct {
	Pixels  [Size]byte
	Palette Palette
	rgba    [Size * 4]byte // pre-allocated RGBA conversion buffer
}

// NewFramebuffer creates a framebuffer with the given palette.
func NewFramebuffer(pal Palette) *Framebuffer {
	return &Framebuffer{Palette: pal}
}

// Clear fills the entire buffer with a single color index.
func (fb *Framebuffer) Clear(colorIndex byte) {
	for i := range fb.Pixels {
		fb.Pixels[i] = colorIndex
	}
}

// SetPixel writes a palette index at (x, y). No bounds checking for speed.
func (fb *Framebuffer) SetPixel(x, y int, colorIndex byte) {
	fb.Pixels[y*Width+x] = colorIndex
}

// SetPixelSafe writes a palette index at (x, y) with bounds checking.
func (fb *Framebuffer) SetPixelSafe(x, y int, colorIndex byte) {
	if x >= 0 && x < Width && y >= 0 && y < Height {
		fb.Pixels[y*Width+x] = colorIndex
	}
}

// GetPixel reads the palette index at (x, y). No bounds checking.
func (fb *Framebuffer) GetPixel(x, y int) byte {
	return fb.Pixels[y*Width+x]
}

// HLine draws a horizontal line from (x0, y) to (x1, y).
func (fb *Framebuffer) HLine(x0, x1, y int, colorIndex byte) {
	if y < 0 || y >= Height {
		return
	}
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if x0 < 0 {
		x0 = 0
	}
	if x1 >= Width {
		x1 = Width - 1
	}
	offset := y * Width
	for x := x0; x <= x1; x++ {
		fb.Pixels[offset+x] = colorIndex
	}
}

// VLine draws a vertical line from (x, y0) to (x, y1).
func (fb *Framebuffer) VLine(x, y0, y1 int, colorIndex byte) {
	if x < 0 || x >= Width {
		return
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	if y0 < 0 {
		y0 = 0
	}
	if y1 >= Height {
		y1 = Height - 1
	}
	for y := y0; y <= y1; y++ {
		fb.Pixels[y*Width+x] = colorIndex
	}
}

// FillRect fills a rectangle with a color index.
func (fb *Framebuffer) FillRect(x0, y0, w, h int, colorIndex byte) {
	for y := y0; y < y0+h; y++ {
		fb.HLine(x0, x0+w-1, y, colorIndex)
	}
}

// RGBA converts the indexed framebuffer to RGBA bytes using the current palette.
// Returns a slice suitable for ebiten.Image.WritePixels().
func (fb *Framebuffer) RGBA() []byte {
	for i, idx := range fb.Pixels {
		c := fb.Palette[idx]
		off := i * 4
		fb.rgba[off] = c.R
		fb.rgba[off+1] = c.G
		fb.rgba[off+2] = c.B
		fb.rgba[off+3] = c.A
	}
	return fb.rgba[:]
}

// SetPalette replaces the current palette.
func (fb *Framebuffer) SetPalette(pal Palette) {
	fb.Palette = pal
}

// SetPaletteColor sets a single palette entry.
func (fb *Framebuffer) SetPaletteColor(index byte, c color.RGBA) {
	fb.Palette[index] = c
}

// CopyFrom copies another framebuffer's pixels (not palette) into this one.
func (fb *Framebuffer) CopyFrom(src *Framebuffer) {
	copy(fb.Pixels[:], src.Pixels[:])
}

// FadeToBlack gradually darkens the palette over duration.
// Returns a channel that closes when fade is complete.
func (fb *Framebuffer) FadeToBlack(duration time.Duration) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		startTime := time.Now()
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		for {
			elapsed := time.Since(startTime)
			if elapsed >= duration {
				for i := range fb.Palette {
					fb.Palette[i] = color.RGBA{0, 0, 0, 255}
				}
				close(done)
				return
			}
			t := 1.0 - float64(elapsed)/float64(duration)
			for i := range fb.Palette {
				c := fb.Palette[i]
				fb.Palette[i] = color.RGBA{
					R: uint8(float64(c.R) * t),
					G: uint8(float64(c.G) * t),
					B: uint8(float64(c.B) * t),
					A: c.A,
				}
			}
			<-ticker.C
		}
	}()
	return done
}
