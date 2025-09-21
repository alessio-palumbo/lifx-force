package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	var (
		h, s           float64 = 180, 100
		hsbk0                  = &HSBK{Hue: &h, Saturation: &s}
		invalidPattern         = FingerPattern{1, 2, 3, 4, 5}
		handClosed             = FingerPattern{0, 0, 0, 0, 0}
	)

	testCases := map[string]struct {
		cfg     *Config
		wantErr string
	}{
		"invalid transition_ms": {
			cfg: &Config{
				General: General{TransitionMs: 0},
			},
			wantErr: "general.transition_ms must be > 0",
		},
		"invalid logging level": {
			cfg: &Config{
				General: General{TransitionMs: 1},
				Logging: Logging{Level: "panic"},
			},
			wantErr: "logging.level must be one of debug, info, warn, error",
		},
		"invalid tracking: frame_skip": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 0},
			},
			wantErr: "tracking.frame_skip must be > 0",
		},
		"invalid tracking: buffer_size": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: -1},
			},
			wantErr: "tracking.buffer_size must be > 0",
		},
		"invalid tracking: gesture_threshold - too small": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0},
			},
			wantErr: "tracking.gesture_threshold must be > 0.0",
		},
		"invalid tracking: gesture_threshold - too large": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 1.2},
			},
			wantErr: "tracking.gesture_threshold must be <= 1.0",
		},
		"invalid gesture binding: gesture": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Gesture: "swoop"},
				},
			},
			wantErr: "bindings[0]: invalid gesture: swoop",
		},
		"invalid gesture binding: selector": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Gesture: GestureSwipeLeft, Selector: Selector{Type: "serial"}},
				},
			},
			wantErr: "bindings[0]: invalid serial value: expected 12 hex chars (6 bytes), got 0",
		},
		"invalid gesture binding: action required": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Gesture: GestureSwipeLeft, Selector: Selector{Type: "all"}},
				},
			},
			wantErr: "bindings[0]: action is required",
		},
		"invalid gesture binding: invalid action": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Gesture: GestureSwipeLeft, Selector: Selector{Type: "all"}, Action: "Unknown"},
				},
			},
			wantErr: "bindings[0]: invalid action: Unknown",
		},
		"invalid gesture binding: action missing args": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Gesture: GestureSwipeLeft, Selector: Selector{Type: "all"}, Action: ActionPowerSetColor},
				},
			},
			wantErr: "bindings[0]: hsbk must be set for action set_color",
		},
		"invalid gesture binding: emtpy HSBK for action": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Gesture: GestureSwipeLeft, Selector: Selector{Type: "all"}, Action: ActionPowerSetColor, HSBK: &HSBK{}},
				},
			},
			wantErr: "bindings[0]: hsbk must be set for action set_color",
		},
		"invalid finger binding: fingers": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Pattern: &invalidPattern},
				},
			},
			wantErr: "bindings[0]: pattern should only contain 0s & 1s",
		},
		"invalid finger binding: selector": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Pattern: &handClosed, Selector: Selector{Type: "serial"}},
				},
			},
			wantErr: "bindings[0]: invalid serial value: expected 12 hex chars (6 bytes), got 0",
		},
		"invalid finger binding: action required": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Pattern: &handClosed, Selector: Selector{Type: "all"}},
				},
			},
			wantErr: "bindings[0]: action is required",
		},
		"invalid finger binding: invalid action": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Pattern: &handClosed, Selector: Selector{Type: "all"}, Action: "Unknown"},
				},
			},
			wantErr: "bindings[0]: invalid action: Unknown",
		},
		"invalid finger binding: action missing args": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Pattern: &handClosed, Selector: Selector{Type: "all"}, Action: ActionPowerSetColor},
				},
			},
			wantErr: "bindings[0]: hsbk must be set for action set_color",
		},
		"invalid finger binding: emtpy HSBK for action": {
			cfg: &Config{
				General:  General{TransitionMs: 1},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
				Bindings: []Binding{
					{Pattern: &handClosed, Selector: Selector{Type: "all"}, Action: ActionPowerSetColor, HSBK: &HSBK{}},
				},
			},
			wantErr: "bindings[0]: hsbk must be set for action set_color",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.EqualError(t, tc.cfg.Validate(), tc.wantErr)
		})
	}

	cfg0 := &Config{
		General:  General{TransitionMs: 1},
		Logging:  Logging{Level: "info"},
		Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
		Bindings: []Binding{
			{Gesture: GestureSwipeLeft, Selector: Selector{Type: "all"}, Action: ActionPowerOff},
			{Pattern: &handClosed, Selector: Selector{Type: "all"}, Action: ActionPowerSetColor, HSBK: hsbk0},
		},
	}
	assert.NoError(t, cfg0.Validate())
}
