package sync

import (
	"encoding/json"
	"fmt"
	"os"
)

type CueFile struct {
	Effects []string `json:"effects"`
	Cues    []CueDef `json:"cues"`
}

type CueDef struct {
	Order       int    `json:"order"`
	Row         int    `json:"row"`
	Effect      string `json:"effect"`
	Transition  string `json:"transition"`
	FadeDur     float64 `json:"fade_dur"`
}

func LoadCueFile(path string) (*Timeline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read cue file: %w", err)
	}

	var cf CueFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("failed to parse cue file: %w", err)
	}

	effectMap := make(map[string]int)
	for i, name := range cf.Effects {
		effectMap[name] = i
	}

	cues := make([]Cue, len(cf.Cues))
	for i, cd := range cf.Cues {
		idx, ok := effectMap[cd.Effect]
		if !ok {
			return nil, fmt.Errorf("unknown effect: %s", cd.Effect)
		}
		transition := cd.Transition
		if transition == "" {
			transition = "cut"
		}
		fadeDur := cd.FadeDur
		if fadeDur <= 0 {
			fadeDur = 1.0
		}
		cues[i] = Cue{
			Pos:       Position{Order: cd.Order, Row: cd.Row},
			EffectIdx: idx,
			Transition: transition,
			FadeDur:   fadeDur,
		}
	}

	return NewTimeline(cues), nil
}
