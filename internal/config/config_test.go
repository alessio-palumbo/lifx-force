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

		h0, h1    float64 = 240, 0
		p0        float64 = 100
		defaultMs         = 1
		userCfg0          = &Config{
			General: General{TransitionMs: 10},
			Logging: Logging{Level: "info", File: "lifx-force.log"},
			GestureBindings: []GestureBinding{
				{
					Gesture:  "swipe_left",
					Action:   "set_color",
					Selector: Selector{Type: SelectorTypeSerial, Value: "d073d5000000", Serial: serial0},
					HSBK: &HSBK{
						Hue: &h0, Saturation: &p0, Brightness: &p0,
					},
				},
				{
					Gesture:  "swipe_right",
					Action:   "set_color",
					Selector: Selector{Type: SelectorTypeSerial, Value: "d073d5000000", Serial: serial0},
					HSBK: &HSBK{
						Hue: &h1, Saturation: &p0, Brightness: &p0,
					},
				},
			},
			FingerBindings: []FingerBinding{
				{
					Pattern:  FingerPattern{1, 1, 1, 1, 1},
					Action:   "power_on",
					Selector: Selector{Type: SelectorTypeAll},
				},
				{
					Pattern:  FingerPattern{0, 0, 0, 0, 0},
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
				General: General{TransitionMs: defaultMs},
				Logging: Logging{Level: "info"},
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
}
