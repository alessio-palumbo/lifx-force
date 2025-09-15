package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	tempFilePathEmpty := filepath.Join(tempDir, "config.toml")

	var (
		h0, h1    uint16 = 240, 0
		p0        uint16 = 100
		defaultMs        = 1
		levelInfo        = "info"
		logFile          = "lifx-force.log"
	)
	testCases := map[string]struct {
		userConfigPath string
		want           *Config
	}{
		"no user config": {
			userConfigPath: tempFilePathEmpty,
			want: &Config{
				General: general{DefaultDurationMs: &defaultMs},
				Logging: logging{Level: &levelInfo, File: &logFile},
				GestureBindings: []GestureBinding{
					{
						Gesture:  "swipe_left",
						Action:   "set_color",
						Selector: Selector{Type: SelectorTypeSerial, ID: "d073d5000000"},
						HSBK: &HSBK{
							Hue: &h0, Saturation: &p0, Brightness: &p0,
						},
					},
					{
						Gesture:  "swipe_right",
						Action:   "set_color",
						Selector: Selector{Type: SelectorTypeSerial, ID: "d073d5000000"},
						HSBK: &HSBK{
							Hue: &h1, Saturation: &p0, Brightness: &p0,
						},
					},
				},
				FingerBindings: []FingerBinding{
					{
						Pattern:  []int{1, 1, 1, 1, 1},
						Action:   "turn_on",
						Selector: Selector{Type: SelectorTypeAll},
					},
					{
						Pattern:  []int{0, 0, 0, 0, 0},
						Action:   "turn_off",
						Selector: Selector{Type: SelectorTypeAll},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := LoadConfig(tc.userConfigPath)
			assert.NoError(t, err)
			assert.Equal(t, got, tc.want)
			assert.FileExists(t, tc.userConfigPath)
		})
	}
}

func Test_merge(t *testing.T) {
	var (
		h0, h1                uint16 = 240, 0
		p0                    uint16 = 100
		defaultMs, userMs            = 1, 100
		levelInfo, debugLevel        = "info", "debug"
		logFile                      = "lifx-force.log"
		noLogFile                    = ""
	)
	testCases := map[string]struct {
		baseConfig *Config
		userConfig *Config
		want       *Config
	}{
		"no user config": {
			baseConfig: &Config{
				General: general{DefaultDurationMs: &defaultMs},
				Logging: logging{Level: &levelInfo, File: &logFile},
				GestureBindings: []GestureBinding{
					{
						Gesture:  "swipe_left",
						Action:   "set_color",
						Selector: Selector{Type: SelectorTypeAll},
						HSBK: &HSBK{
							Hue: &h0, Saturation: &p0, Brightness: &p0,
						},
					},
					{
						Gesture:  "swipe_right",
						Action:   "set_color",
						Selector: Selector{Type: SelectorTypeAll},
						HSBK: &HSBK{
							Hue: &h1, Saturation: &p0, Brightness: &p0,
						},
					},
				},
				FingerBindings: []FingerBinding{
					{
						Pattern:  []int{1, 1, 1, 1, 1},
						Action:   "turn_on",
						Selector: Selector{Type: SelectorTypeAll},
					},
				},
			},
			userConfig: &Config{},
			want: &Config{
				General: general{DefaultDurationMs: &defaultMs},
				Logging: logging{Level: &levelInfo, File: &logFile},
				GestureBindings: []GestureBinding{
					{
						Gesture:  "swipe_left",
						Action:   "set_color",
						Selector: Selector{Type: SelectorTypeAll},
						HSBK: &HSBK{
							Hue: &h0, Saturation: &p0, Brightness: &p0,
						},
					},
					{
						Gesture:  "swipe_right",
						Action:   "set_color",
						Selector: Selector{Type: SelectorTypeAll},
						HSBK: &HSBK{
							Hue: &h1, Saturation: &p0, Brightness: &p0,
						},
					},
				},
				FingerBindings: []FingerBinding{
					{
						Pattern:  []int{1, 1, 1, 1, 1},
						Action:   "turn_on",
						Selector: Selector{Type: SelectorTypeAll},
					},
				},
			},
		},
		"has user config": {
			baseConfig: &Config{
				General: general{DefaultDurationMs: &defaultMs},
				Logging: logging{Level: &levelInfo, File: &logFile},
				GestureBindings: []GestureBinding{
					{
						Gesture:  "swipe_left",
						Action:   "set_brightness",
						Selector: Selector{Type: SelectorTypeAll},
						HSBK:     &HSBK{Brightness: &p0},
					},
				},
				FingerBindings: []FingerBinding{
					{
						Pattern:  []int{1, 1, 1, 1, 1},
						Action:   "turn_on",
						Selector: Selector{Type: SelectorTypeAll},
					},
				},
			},
			userConfig: &Config{
				General: general{DefaultDurationMs: &userMs},
				Logging: logging{Level: &debugLevel, File: &noLogFile},
				FingerBindings: []FingerBinding{
					{
						Pattern:  []int{0, 0, 0, 0, 0},
						Action:   "turn_off",
						Selector: Selector{Type: SelectorTypeAll},
					},
				},
			},
			want: &Config{
				General: general{DefaultDurationMs: &userMs},
				Logging: logging{Level: &debugLevel, File: &noLogFile},
				GestureBindings: []GestureBinding{
					{
						Gesture:  "swipe_left",
						Action:   "set_brightness",
						Selector: Selector{Type: SelectorTypeAll},
						HSBK:     &HSBK{Brightness: &p0},
					},
				},
				FingerBindings: []FingerBinding{
					{
						Pattern:  []int{0, 0, 0, 0, 0},
						Action:   "turn_off",
						Selector: Selector{Type: SelectorTypeAll},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := merge(tc.baseConfig, tc.userConfig)
			assert.NoError(t, err)
			assert.Equal(t, tc.baseConfig, tc.want)
		})
	}
}
