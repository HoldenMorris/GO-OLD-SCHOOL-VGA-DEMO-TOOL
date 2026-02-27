package music

/*
#cgo pkg-config: libxmp
#include <xmp.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// MaxChannels is the maximum number of tracker channels we expose.
const MaxChannels = 64

// FrameInfo holds the current playback state, suitable for syncing visuals.
type FrameInfo struct {
	Order      int     // Current position in the order list
	Pattern    int     // Current pattern number
	Row        int     // Current row in the pattern
	NumRows    int     // Total rows in the current pattern
	Frame      int     // Current frame (tick) within the row
	Speed      int     // Ticks per row
	BPM        int     // Beats per minute
	TimeMs     int     // Elapsed time in milliseconds
	TotalTimeMs int    // Total module time in milliseconds
	LoopCount  int     // Number of times the module has looped
	ChannelVol [MaxChannels]int // Per-channel volume (0-255)
	NumChannels int    // Number of active channels

	// Derived sync helpers
	BeatProgress float64 // 0.0-1.0 progress through current row
}

// Context wraps an xmp_context for tracker module playback.
type Context struct {
	ctx C.xmp_context
}

// NewContext creates a new libxmp player context.
func NewContext() *Context {
	return &Context{ctx: C.xmp_create_context()}
}

// Close frees the libxmp context.
func (c *Context) Close() {
	if c.ctx != nil {
		C.xmp_free_context(c.ctx)
		c.ctx = nil
	}
}

// LoadModule loads a tracker module from a file path.
func (c *Context) LoadModule(path string) error {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	ret := C.xmp_load_module(c.ctx, cpath)
	if ret < 0 {
		return fmt.Errorf("xmp_load_module failed: %d", ret)
	}
	return nil
}

// LoadModuleFromMemory loads a tracker module from a byte slice.
func (c *Context) LoadModuleFromMemory(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty module data")
	}
	ret := C.xmp_load_module_from_memory(c.ctx, unsafe.Pointer(&data[0]), C.long(len(data)))
	if ret < 0 {
		return fmt.Errorf("xmp_load_module_from_memory failed: %d", ret)
	}
	return nil
}

// ReleaseModule releases the currently loaded module.
func (c *Context) ReleaseModule() {
	C.xmp_release_module(c.ctx)
}

// StartPlayer starts playback at the given sample rate.
func (c *Context) StartPlayer(sampleRate int) error {
	ret := C.xmp_start_player(c.ctx, C.int(sampleRate), 0)
	if ret < 0 {
		return fmt.Errorf("xmp_start_player failed: %d", ret)
	}
	return nil
}

// EndPlayer stops playback.
func (c *Context) EndPlayer() {
	C.xmp_end_player(c.ctx)
}

// PlayFrame renders one frame of audio. Returns false when the module ends.
func (c *Context) PlayFrame() bool {
	ret := C.xmp_play_frame(c.ctx)
	return ret == 0
}

// GetFrameInfo fills a FrameInfo struct with current playback state.
func (c *Context) GetFrameInfo() FrameInfo {
	var fi C.struct_xmp_frame_info
	C.xmp_get_frame_info(c.ctx, &fi)

	info := FrameInfo{
		Order:       int(fi.pos),
		Pattern:     int(fi.pattern),
		Row:         int(fi.row),
		NumRows:     int(fi.num_rows),
		Frame:       int(fi.frame),
		Speed:       int(fi.speed),
		BPM:         int(fi.bpm),
		TimeMs:      int(fi.time),
		TotalTimeMs: int(fi.total_time),
		LoopCount:   int(fi.loop_count),
	}

	// Extract per-channel volumes
	numCh := int(fi.virt_used)
	if numCh > MaxChannels {
		numCh = MaxChannels
	}
	info.NumChannels = numCh
	for i := 0; i < numCh; i++ {
		info.ChannelVol[i] = int(fi.channel_info[i].volume)
	}

	// Beat progress: how far through the current row (0.0 to 1.0)
	if info.Speed > 0 {
		info.BeatProgress = float64(info.Frame) / float64(info.Speed)
	}

	return info
}

// GetBuffer returns the PCM audio buffer for the current frame.
// The buffer contains interleaved 16-bit signed samples (stereo).
func (c *Context) GetBuffer() []byte {
	var fi C.struct_xmp_frame_info
	C.xmp_get_frame_info(c.ctx, &fi)
	size := int(fi.buffer_size)
	if size <= 0 {
		return nil
	}
	return C.GoBytes(fi.buffer, C.int(size))
}

// SetPosition seeks to a specific order position.
func (c *Context) SetPosition(pos int) error {
	ret := C.xmp_set_position(c.ctx, C.int(pos))
	if ret < 0 {
		return fmt.Errorf("xmp_set_position failed: %d", ret)
	}
	return nil
}
