package vga

import "image/color"

// Palette is a 256-color lookup table, matching VGA Mode 13h.
type Palette [256]color.RGBA

// DefaultPalette returns the standard VGA 256-color palette (Mode 13h default).
// Colors 0-15: standard CGA colors
// Colors 16-231: 6x6x6 color cube
// Colors 232-255: grayscale ramp
func DefaultPalette() Palette {
	var p Palette

	// CGA colors (0-15)
	cga := [16]color.RGBA{
		{0, 0, 0, 255},       // 0: black
		{0, 0, 170, 255},     // 1: blue
		{0, 170, 0, 255},     // 2: green
		{0, 170, 170, 255},   // 3: cyan
		{170, 0, 0, 255},     // 4: red
		{170, 0, 170, 255},   // 5: magenta
		{170, 85, 0, 255},    // 6: brown
		{170, 170, 170, 255}, // 7: light gray
		{85, 85, 85, 255},    // 8: dark gray
		{85, 85, 255, 255},   // 9: light blue
		{85, 255, 85, 255},   // 10: light green
		{85, 255, 255, 255},  // 11: light cyan
		{255, 85, 85, 255},   // 12: light red
		{255, 85, 255, 255},  // 13: light magenta
		{255, 255, 85, 255},  // 14: yellow
		{255, 255, 255, 255}, // 15: white
	}
	copy(p[:16], cga[:])

	// 6x6x6 color cube (16-231)
	i := 16
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				p[i] = color.RGBA{
					R: uint8(r * 51),
					G: uint8(g * 51),
					B: uint8(b * 51),
					A: 255,
				}
				i++
			}
		}
	}

	// Grayscale ramp (232-255)
	for j := 0; j < 24; j++ {
		v := uint8(8 + j*10)
		p[232+j] = color.RGBA{v, v, v, 255}
	}

	return p
}

// FirePalette returns a black → red → yellow → white gradient for fire effects.
func FirePalette() Palette {
	var p Palette
	for i := 0; i < 256; i++ {
		var r, g, b uint8
		switch {
		case i < 64:
			r = uint8(i * 4)
		case i < 128:
			r = 255
			g = uint8((i - 64) * 4)
		case i < 192:
			r = 255
			g = 255
			b = uint8((i - 128) * 4)
		default:
			r, g, b = 255, 255, 255
		}
		p[i] = color.RGBA{r, g, b, 255}
	}
	return p
}

// GradientPalette creates a smooth gradient between two colors across all 256 entries.
func GradientPalette(from, to color.RGBA) Palette {
	var p Palette
	for i := 0; i < 256; i++ {
		t := float64(i) / 255.0
		p[i] = color.RGBA{
			R: uint8(float64(from.R) + t*float64(int(to.R)-int(from.R))),
			G: uint8(float64(from.G) + t*float64(int(to.G)-int(from.G))),
			B: uint8(float64(from.B) + t*float64(int(to.B)-int(from.B))),
			A: 255,
		}
	}
	return p
}

// PlasmaPalette returns a colorful cycling palette suitable for plasma effects.
func PlasmaPalette() Palette {
	var p Palette
	for i := 0; i < 256; i++ {
		t := float64(i) / 256.0 * 3.14159265 * 2.0
		r := uint8(128.0 + 127.0*sin(t))
		g := uint8(128.0 + 127.0*sin(t+2.094))       // +2π/3
		b := uint8(128.0 + 127.0*sin(t+4.189))       // +4π/3
		p[i] = color.RGBA{r, g, b, 255}
	}
	return p
}

// sin is a quick inline to avoid importing math just for this.
func sin(x float64) float64 {
	// Normalize to [-π, π]
	for x > 3.14159265 {
		x -= 6.28318530
	}
	for x < -3.14159265 {
		x += 6.28318530
	}
	// Bhaskara I approximation — good enough for palette generation
	if x < 0 {
		return -sin(-x)
	}
	return (16*x*(3.14159265-x))/(49.348-(4*x*(3.14159265-x)))
}
