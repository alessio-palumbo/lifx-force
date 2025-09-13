package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"

	_ "embed"
)

//go:embed default.toml
var defaultConfigData []byte

type SelectorType string

const (
	SelectorTypeAll    SelectorType = "all"
	SelectorTypeSerial SelectorType = "serial"
)

type Config struct {
	General general `toml:"general"`

	Logging logging `toml:"logging"`

	GestureBindings []GestureBinding `toml:"gesture_bindings"`
	FingerBindings  []FingerBinding  `toml:"finger_bindings"`
}

type general struct {
	DefaultDurationMs int `toml:"default_duration_ms"`
}

type logging struct {
	Level string `toml:"level"`
	File  string `toml:"file"`
}

type GestureBinding struct {
	Gesture  string   `toml:"gesture"`
	Action   string   `toml:"action"`
	Selector Selector `toml:"selector"`
	HSBK     *HSBK    `toml:"hsbk,omitempty"`
}

type FingerBinding struct {
	Pattern  []int    `toml:"pattern"`
	Action   string   `toml:"action"`
	Selector Selector `toml:"selector"`
	HSBK     *HSBK    `toml:"hsbk,omitempty"`
}

type Selector struct {
	Type SelectorType
	ID   string
}

type HSBK struct {
	Hue        *uint16 `toml:"hue"`
	Saturation *uint16 `toml:"saturation"`
	Brightness *uint16 `toml:"brightness"`
	Kelvin     *uint16 `toml:"kelvin"`
}

func LoadConfig() (*Config, error) {
	userCfg, err := loadUserConfig()
	if err != nil {
		return nil, err
	}

	return loadConfig(userCfg)
}

func loadConfig(userCfg *Config) (*Config, error) {
	baseCfg := &Config{}
	if err := toml.Unmarshal(defaultConfigData, baseCfg); err != nil {
		return nil, err
	}
	if userCfg == nil {
		return baseCfg, nil
	}

	if err := merge(baseCfg, userCfg); err != nil {
		return nil, err
	}
	return baseCfg, nil
}

// loadUserConfig loads the user config if it exists.
func loadUserConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	userConfigPath := filepath.Join(homeDir, ".lifx-force", "config.toml")
	cfg, err := readConfigFile(userConfigPath)
	if err != nil {
		// Do not error if user config is not supplied.
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return cfg, nil
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
