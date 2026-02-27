package effects

import (
	"github.com/holden/vga-go/internal/music"
	"github.com/holden/vga-go/internal/vga"
)

// Effect is the interface all demo effects implement.
type Effect interface {
	// Init is called once when the effect becomes active.
	Init(fb *vga.Framebuffer)

	// Update is called each frame with the current music sync state.
	// dt is the delta time in seconds since the last frame.
	Update(dt float64, sync music.FrameInfo)

	// Draw renders the effect into the framebuffer.
	Draw(fb *vga.Framebuffer)
}
