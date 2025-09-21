package config

import (
	"path/filepath"
	"testing"

	"github.com/alessio-palumbo/lifxlan-go/pkg/device"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	var (
		tempDir            = t.TempDir()
		tempFilePathEmpty  = filepath.Join(tempDir, "config-empty.toml")
		tempFilePathEdited = filepath.Join(tempDir, "config-edited.toml")
		serial0, _         = device.SerialFromHex("d073d5000000")

		handClosed = FingerPattern{0, 0, 0, 0, 0}
		handOpen   = FingerPattern{1, 1, 1, 1, 1}

		h0, h1    float64 = 240, 0
		p0        float64 = 100
		defaultMs         = 1
		userCfg0          = &Config{
			General:  General{TransitionMs: 10},
			Logging:  Logging{Level: "info", File: "lifx-force.log"},
			Tracking: Tracking{FrameSkip: 1, BufferSize: 8, GestureThreshold: 0.3},
			Bindings: []Binding{
				{
					Gesture:  GestureSwipeLeft,
					Action:   "set_color",
					Selector: Selector{Type: SelectorTypeSerial, Value: "d073d5000000", Serial: serial0},
					HSBK:     &HSBK{Hue: &h0, Saturation: &p0, Brightness: &p0},
				},
				{
					Gesture:  GestureSwipeRight,
					Action:   "set_color",
					Selector: Selector{Type: SelectorTypeLabel, Value: "lamp"},
					HSBK:     &HSBK{Hue: &h1, Saturation: &p0, Brightness: &p0},
				},
				{
					Gesture:  GestureSwipeUp,
					Action:   "set_color",
					Selector: Selector{Type: SelectorTypeLabel, Value: "desk"},
					HSBK:     &HSBK{Hue: &h1, Saturation: &p0, Brightness: &p0},
				},
				{
					Gesture:  GestureSwipeDown,
					Action:   "set_color",
					Selector: Selector{Type: SelectorTypeLabel, Value: "left bulb"},
					HSBK:     &HSBK{Hue: &h1, Saturation: &p0, Brightness: &p0},
				},
				{
					Gesture:  GestureExpand,
					Action:   "power_off",
					Selector: Selector{Type: SelectorTypeLabel, Value: "right_bulb"},
				},
				{
					Gesture:  GestureContract,
					Action:   "power_on",
					Selector: Selector{Type: SelectorTypeLabel, Value: "right_bulb"},
				},
				{
					Gesture:  GesturePushDown,
					Action:   "power_off",
					Selector: Selector{Type: SelectorTypeGroup, Value: "living room"},
				},
				{
					Gesture:  GesturePullUp,
					Action:   "power_on",
					Selector: Selector{Type: SelectorTypeGroup, Value: "living room"},
				},
				{
					Pattern:  &handOpen,
					Action:   "power_on",
					Selector: Selector{Type: SelectorTypeAll},
				},
				{
					Pattern:  &handClosed,
					Action:   "power_off",
					Selector: Selector{Type: SelectorTypeAll},
				},
			},
		}
	)

	if err := createConfigFile(userCfg0, tempFilePathEdited); err != nil {
		t.Fatal(err)
	}

	testCases := map[string]struct {
		userConfigPath string
		want           *Config
	}{
		"no user config": {
			userConfigPath: tempFilePathEmpty,
			want: &Config{
				General:  General{TransitionMs: defaultMs},
				Logging:  Logging{Level: "info"},
				Tracking: Tracking{FrameSkip: 1, BufferSize: 5, GestureThreshold: 0.1},
			},
		},
		"with user config": {
			userConfigPath: tempFilePathEdited,
			want:           userCfg0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := LoadConfig(tc.userConfigPath)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.FileExists(t, tc.userConfigPath)
		})
	}

	// Unhappy Path
	userCfg1 := userCfg0
	userCfg1.General.TransitionMs = -1
	tempFilePathEditedInvalid := filepath.Join(tempDir, "config-edited-invalid.toml")

	if err := createConfigFile(userCfg1, tempFilePathEditedInvalid); err != nil {
		t.Fatal(err)
	}
	got, err := LoadConfig(tempFilePathEditedInvalid)
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.FileExists(t, tempFilePathEditedInvalid)
}
