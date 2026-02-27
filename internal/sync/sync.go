package sync

import "github.com/holden/vga-go/internal/music"

// Position identifies a point in the tracker timeline.
type Position struct {
	Order int // Position in the order list
	Row   int // Row within the pattern (-1 means any row)
}

// Cue is a trigger point in the demo timeline.
type Cue struct {
	Pos       Position
	EffectIdx int    // Index of the effect to activate
	Transition string // "cut", "fade", "crossfade"
	FadeDur   float64 // Duration of fade in seconds (for fade/crossfade)
}

// Timeline holds the ordered list of cues for the demo.
type Timeline struct {
	Cues []Cue
}

// NewTimeline creates a timeline from a list of cues.
func NewTimeline(cues []Cue) *Timeline {
	return &Timeline{Cues: cues}
}

// ActiveCue returns the cue that should be active at the given position.
// Returns the index of the matching cue, or -1 if none match yet.
func (t *Timeline) ActiveCue(info music.FrameInfo) int {
	best := -1
	for i, cue := range t.Cues {
		if info.Order > cue.Pos.Order ||
			(info.Order == cue.Pos.Order && info.Row >= cue.Pos.Row) {
			best = i
		}
	}
	return best
}

// BeatPulse returns a 0.0-1.0 value that peaks at 1.0 on each beat (row 0 of each beat)
// and decays to 0.0 by the next beat. Useful for reactive visuals.
func BeatPulse(info music.FrameInfo) float64 {
	if info.Speed <= 0 {
		return 0
	}
	return 1.0 - info.BeatProgress
}

// RowPulse returns 1.0 on the first frame of each row, 0.0 otherwise.
func RowPulse(info music.FrameInfo) float64 {
	if info.Frame == 0 {
		return 1.0
	}
	return 0.0
}

// MaxChannelVolume returns the highest volume across all active channels (0.0-1.0).
func MaxChannelVolume(info music.FrameInfo) float64 {
	max := 0
	for i := 0; i < info.NumChannels; i++ {
		if info.ChannelVol[i] > max {
			max = info.ChannelVol[i]
		}
	}
	return float64(max) / 255.0
}
