package sync

import (
	"github.com/holden/vga-go/internal/effects"
	"github.com/holden/vga-go/internal/music"
	"github.com/holden/vga-go/internal/vga"
)

// Sequencer manages effect transitions driven by the music timeline.
type Sequencer struct {
	effects  []effects.Effect
	timeline *Timeline

	currentIdx int     // index into timeline.Cues
	activeIdx  int     // effects[] index of the active effect
	prevIdx    int     // effects[] index of the previous effect (for crossfade)
	fadeAlpha  float64 // 0.0 = previous, 1.0 = current (for transitions)
	fadeDur    float64 // total fade duration
	fadeTimer  float64 // elapsed fade time
	fading     bool

	initialized map[int]bool // tracks which effects have been Init'd
}

// NewSequencer creates a sequencer with the given effects and timeline.
func NewSequencer(efx []effects.Effect, tl *Timeline) *Sequencer {
	return &Sequencer{
		effects:     efx,
		timeline:    tl,
		currentIdx:  -1,
		activeIdx:   0,
		prevIdx:     -1,
		fadeAlpha:    1.0,
		initialized: make(map[int]bool),
	}
}

// Update advances the sequencer based on current music state.
func (s *Sequencer) Update(dt float64, info music.FrameInfo, fb *vga.Framebuffer) {
	// Check timeline for cue changes
	cueIdx := s.timeline.ActiveCue(info)
	if cueIdx >= 0 && cueIdx != s.currentIdx {
		cue := s.timeline.Cues[cueIdx]
		s.currentIdx = cueIdx

		newEffect := cue.EffectIdx
		if newEffect != s.activeIdx && newEffect >= 0 && newEffect < len(s.effects) {
			// Initialize if first time
			if !s.initialized[newEffect] {
				s.effects[newEffect].Init(fb)
				s.initialized[newEffect] = true
			}

			switch cue.Transition {
			case "fade", "crossfade":
				s.prevIdx = s.activeIdx
				s.activeIdx = newEffect
				s.fadeDur = cue.FadeDur
				s.fadeTimer = 0
				s.fadeAlpha = 0
				s.fading = true
			default: // "cut"
				s.activeIdx = newEffect
				s.fading = false
				s.fadeAlpha = 1.0
				// Apply the new effect's palette immediately
				s.effects[newEffect].Init(fb)
			}
		}
	}

	// Advance fade
	if s.fading {
		s.fadeTimer += dt
		if s.fadeDur > 0 {
			s.fadeAlpha = s.fadeTimer / s.fadeDur
		} else {
			s.fadeAlpha = 1.0
		}
		if s.fadeAlpha >= 1.0 {
			s.fadeAlpha = 1.0
			s.fading = false
		}
	}

	// Update active effect(s)
	if s.activeIdx >= 0 && s.activeIdx < len(s.effects) {
		s.effects[s.activeIdx].Update(dt, info)
	}
	if s.fading && s.prevIdx >= 0 && s.prevIdx < len(s.effects) {
		s.effects[s.prevIdx].Update(dt, info)
	}
}

// Draw renders the current effect(s) into the framebuffer.
func (s *Sequencer) Draw(fb *vga.Framebuffer) {
	if !s.fading {
		// Simple case: just draw the active effect
		if s.activeIdx >= 0 && s.activeIdx < len(s.effects) {
			s.effects[s.activeIdx].Draw(fb)
		}
		return
	}

	// Crossfade: blend two framebuffers
	var prevFB vga.Framebuffer
	prevFB.Palette = fb.Palette

	if s.prevIdx >= 0 && s.prevIdx < len(s.effects) {
		s.effects[s.prevIdx].Draw(&prevFB)
	}
	if s.activeIdx >= 0 && s.activeIdx < len(s.effects) {
		s.effects[s.activeIdx].Draw(fb)
	}

	// Alpha blend at the palette index level (simple: pick based on alpha threshold)
	alpha := s.fadeAlpha
	for i := range fb.Pixels {
		if alpha < 0.5 {
			fb.Pixels[i] = prevFB.Pixels[i]
		}
		// When alpha >= 0.5, keep fb.Pixels[i] (new effect) as-is
	}
}

// InitFirst initializes the first effect in the timeline.
func (s *Sequencer) InitFirst(fb *vga.Framebuffer) {
	if len(s.effects) > 0 {
		idx := 0
		if len(s.timeline.Cues) > 0 {
			idx = s.timeline.Cues[0].EffectIdx
		}
		s.activeIdx = idx
		s.effects[idx].Init(fb)
		s.initialized[idx] = true
	}
}
