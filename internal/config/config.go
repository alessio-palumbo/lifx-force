package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/alessio-palumbo/lifxlan-go/pkg/device"

	_ "embed"
)

const (
	defaultTransitionMs = 1

	defaultLogLevel = "info"

	defaultFrameSkip        = 1
	defaultBufferSize       = 5
	defaultGestureThreshold = 0.1
)

type FingerPattern [5]int

type Gesture string

const (
	GestureSwipeLeft  Gesture = "swipe_left"
	GestureSwipeRight Gesture = "swipe_right"
)

type Action string

const (
	ActionPowerOn       Action = "power_on"
	ActionPowerOff      Action = "power_off"
	ActionPowerSetColor Action = "set_color"
)

type SelectorType string

const (
	SelectorTypeAll      SelectorType = "all"
	SelectorTypeLabel    SelectorType = "label"
	SelectorTypeGroup    SelectorType = "group"
	SelectorTypeLocation SelectorType = "location"
	SelectorTypeSerial   SelectorType = "serial"
)

type Config struct {
	General  General  `toml:"general"`
	Logging  Logging  `toml:"logging"`
	Tracking Tracking `toml:"tracking"`

	GestureBindings []GestureBinding `toml:"gesture_bindings"`
	FingerBindings  []FingerBinding  `toml:"finger_bindings"`
}

type General struct {
	TransitionMs int `toml:"transition_ms"`
}

type Tracking struct {
	FrameSkip        int     `toml:"frame_skip"`
	BufferSize       int     `toml:"buffer_size"`
	GestureThreshold float64 `toml:"gesture_threshold"`
}

type Logging struct {
	Level string `toml:"level"`
	File  string `toml:"file"`
}

type GestureBinding struct {
	Gesture  Gesture  `toml:"gesture"`
	Action   Action   `toml:"action"`
	Selector Selector `toml:"selector"`
	HSBK     *HSBK    `toml:"hsbk,omitempty"`
}

type FingerBinding struct {
	Pattern  FingerPattern `toml:"pattern"`
	Action   Action        `toml:"action"`
	Selector Selector      `toml:"selector"`
	HSBK     *HSBK         `toml:"hsbk,omitempty"`
}

type HSBK struct {
	Hue        *float64 `toml:"hue"`
	Saturation *float64 `toml:"saturation"`
	Brightness *float64 `toml:"brightness"`
	Kelvin     *uint16  `toml:"kelvin"`
}

type Selector struct {
	Type  SelectorType `toml:"type"`
	Value string       `toml:"value,omitempty"`
	// Serial is set on unmarshalling when type is SelectorTypeSerial
	Serial device.Serial `toml:"-"`
}

func LoadConfig(userConfigPath string) (*Config, error) {
	baseCfg := newBaseConfig()

	// Create user config based on the default file, if it does not exists.
	if _, err := os.Stat(userConfigPath); errors.Is(err, os.ErrNotExist) {
		if err := createConfigFile(baseCfg, userConfigPath); err != nil {
			return nil, err
		}
		return baseCfg, nil
	}

	userCfg, err := readConfigFile(userConfigPath)
	if err != nil {
		return nil, err
	}

	if err := merge(baseCfg, userCfg); err != nil {
		return nil, err
	}

	if err := baseCfg.Validate(); err != nil {
		return nil, err
	}
	return baseCfg, nil
}

func newBaseConfig() *Config {
	return &Config{
		General: General{TransitionMs: defaultTransitionMs},
		Logging: Logging{Level: defaultLogLevel},
		Tracking: Tracking{
			FrameSkip:        defaultFrameSkip,
			BufferSize:       defaultBufferSize,
			GestureThreshold: defaultGestureThreshold,
		},
	}
}

func createConfigFile(cfg *Config, path string) error {
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal defaults: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, buf, 0644); err != nil {
		return err
	}

	return nil
}

func readConfigFile(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// merge user into base, overriding only non-nil values in user.
// Both base and user must be pointers to structs.
func merge(base, user any) error {
	baseVal := reflect.ValueOf(base)
	userVal := reflect.ValueOf(user)
	if userVal.IsZero() {
		return nil
	}

	if baseVal.Kind() != reflect.Ptr || userVal.Kind() != reflect.Ptr {
		return &reflect.ValueError{Method: "Merge", Kind: baseVal.Kind()}
	}

	baseElem := baseVal.Elem()
	userElem := userVal.Elem()

	if baseElem.Kind() != reflect.Struct || userElem.Kind() != reflect.Struct {
		return &reflect.ValueError{Method: "Merge", Kind: baseElem.Kind()}
	}

	for i := range baseElem.NumField() {
		bf := baseElem.Field(i)
		uf := userElem.Field(i)

		// Only merge exported fields
		if !bf.CanSet() {
			continue
		}

		switch bf.Kind() {
		case reflect.Int, reflect.Float64:
			if !uf.IsZero() {
				bf.Set(uf)
			}
		case reflect.String:
			if uf.Len() > 0 {
				bf.Set(uf)
			}
		case reflect.Ptr:
			if !uf.IsNil() {
				bf.Set(uf)
			}
		case reflect.Slice:
			if uf.Len() > 0 {
				bf.Set(uf)
			}
		case reflect.Struct:
			// Recurse into nested struct
			merge(bf.Addr().Interface(), uf.Addr().Interface())
		}
	}

	return nil
}
