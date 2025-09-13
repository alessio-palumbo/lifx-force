package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_loadConfig(t *testing.T) {
	var (
		h0, h1 uint16 = 240, 0
		p0     uint16 = 100
	)
	testCases := map[string]struct {
		userCfg *Config
		want    *Config
	}{
		"no user config": {
			userCfg: &Config{},
			want: &Config{
				General: general{DefaultDurationMs: 1},
				Logging: logging{Level: "info", File: "lifx-force.log"},
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
						Pattern: []int{1, 1, 1, 1, 1},
						Action:  "turn_on",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := loadConfig(tc.userCfg)
			assert.NoError(t, err)
			assert.Equal(t, got, tc.want)
		})
	}
}
