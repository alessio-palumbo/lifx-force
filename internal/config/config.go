package config

import (
	"errors"
	"fmt"
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
	DefaultDurationMs *int `toml:"default_duration_ms"`
}

type logging struct {
	Level *string `toml:"level"`
	File  *string `toml:"file"`
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

func LoadConfig(userConfigPath string) (*Config, error) {
	baseCfg := &Config{}
	if err := toml.Unmarshal(defaultConfigData, baseCfg); err != nil {
		return nil, err
	}

	// Create user config based on the default file, if it does not exists.
	if _, err := os.Stat(userConfigPath); errors.Is(err, os.ErrNotExist) {
		if err := createUserConfig(baseCfg, userConfigPath); err != nil {
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
	return baseCfg, nil
}

// createUserConfig creates the user config based on the default config.
func createUserConfig(baseCfg *Config, userConfigPath string) error {
	buf, err := toml.Marshal(baseCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal defaults: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(userConfigPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(userConfigPath, buf, 0644); err != nil {
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
		case reflect.Int:
			bf.Set(uf)
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
