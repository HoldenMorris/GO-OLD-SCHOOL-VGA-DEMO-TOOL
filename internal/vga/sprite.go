package vga

type Sprite struct {
	Width  int
	Height int
	Pixels []byte
}

func NewSprite(w, h int) *Sprite {
	return &Sprite{
		Width:  w,
		Height: h,
		Pixels: make([]byte, w*h),
	}
}

func NewSpriteFromData(w, h int, data []byte) *Sprite {
	if len(data) < w*h {
		padded := make([]byte, w*h)
		copy(padded, data)
		return &Sprite{Width: w, Height: h, Pixels: padded}
	}
	return &Sprite{Width: w, Height: h, Pixels: data[:w*h]}
}

func (s *Sprite) SetPixel(x, y int, color byte) {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return
	}
	s.Pixels[y*s.Width+x] = color
}

func (s *Sprite) GetPixel(x, y int) byte {
	if x < 0 || x >= s.Width || y < 0 || y >= s.Height {
		return 0
	}
	return s.Pixels[y*s.Width+x]
}

func (fb *Framebuffer) DrawSprite(x, y int, s *Sprite) {
	for fy := 0; fy < s.Height; fy++ {
		for fx := 0; fx < s.Width; fx++ {
			px := s.GetPixel(fx, fy)
			if px != 0 {
				fb.SetPixel(x+fx, y+fy, px)
			}
		}
	}
}

func (fb *Framebuffer) DrawSpriteScaled(x, y int, s *Sprite, scale int) {
	if scale <= 0 {
		scale = 1
	}
	for fy := 0; fy < s.Height; fy++ {
		for fx := 0; fx < s.Width; fx++ {
			px := s.GetPixel(fx, fy)
			if px != 0 {
				for sy := 0; sy < scale; sy++ {
					for sx := 0; sx < scale; sx++ {
						fb.SetPixel(x+fx*scale+sx, y+fy*scale+sy, px)
					}
				}
			}
		}
	}
}

func CharToSprite(ch byte) *Sprite {
	font := CP437Font[ch]
	s := NewSprite(8, 8)
	for y := 0; y < 8; y++ {
		row := font[y]
		for x := 0; x < 8; x++ {
			if row&(1<<uint(x)) != 0 {
				s.SetPixel(x, y, 255)
			}
		}
	}
	return s
}
