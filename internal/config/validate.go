package config

import (
	"fmt"

	"github.com/alessio-palumbo/lifxlan-go/pkg/device"
)

func (c *Config) Validate() error {
	if c.General.TransitionMs <= 0 {
		return fmt.Errorf("general.transition_ms must be > 0")
	}

	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("logging.level must be one of debug, info, warn, error")
	}

	if err := c.Tracking.Validate(); err != nil {
		return err
	}

	for i := range c.GestureBindings {
		g := &c.GestureBindings[i]
		if err := g.Validate(); err != nil {
			return fmt.Errorf("gesture_bindings[%d]: %w", i, err)
		}
	}

	for i := range c.FingerBindings {
		fb := &c.FingerBindings[i]
		if err := fb.Validate(); err != nil {
			return fmt.Errorf("finger_bindings[%d]: %w", i, err)
		}
	}

	return nil
}

func (t *Tracking) Validate() error {
	if t.FrameSkip <= 0 {
		return fmt.Errorf("tracking.frame_skip must be > 0")
	}
	if t.BufferSize <= 0 {
		return fmt.Errorf("tracking.buffer_size must be > 0")
	}
	if t.GestureThreshold <= 0 {
		return fmt.Errorf("tracking.gesture_threshold must be > 0.0")
	} else if t.GestureThreshold > 1 {
		return fmt.Errorf("tracking.gesture_threshold must be <= 1.0")
	}
	return nil
}

func (g *GestureBinding) Validate() error {
	switch g.Gesture {
	case GestureSwipeLeft, GestureSwipeRight:
	case "":
		return fmt.Errorf("gesture is required")
	default:
		return fmt.Errorf("invalid gesture: %s", g.Gesture)
	}

	if err := g.Selector.Validate(); err != nil {
		return err
	}
	if err := ValidateActionAndArgs(g.Action, g.HSBK); err != nil {
		return err
	}

	return nil
}

func (f *FingerBinding) Validate() error {
	for _, p := range f.Pattern {
		if p != 0 && p != 1 {
			return fmt.Errorf("pattern should only contain 0s & 1s")
		}
	}

	if err := f.Selector.Validate(); err != nil {
		return err
	}
	if err := ValidateActionAndArgs(f.Action, f.HSBK); err != nil {
		return err
	}

	return nil
}

func (s *Selector) Validate() error {
	switch s.Type {
	case SelectorTypeAll:
	case SelectorTypeLabel, SelectorTypeGroup, SelectorTypeLocation:
		if len(s.Value) == 0 {
			return fmt.Errorf("missing selector value for type %q", s.Type)
		}
	case SelectorTypeSerial:
		serial, err := device.SerialFromHex(s.Value)
		if err != nil {
			return fmt.Errorf("invalid serial value: %w", err)
		}
		s.Serial = serial
	default:
		return fmt.Errorf("unknown selector type %q", s.Type)
	}
	return nil
}

func (hsbk *HSBK) Validate() error {
	if hsbk == nil {
		return nil
	}
	if h := hsbk.Hue; h != nil {
		if *h < 0 || *h > 360 {
			return fmt.Errorf("invalid value for hue [%v], must be 0-360", *h)
		}
	}
	if s := hsbk.Saturation; s != nil {
		if *s < 0 || *s > 100 {
			return fmt.Errorf("invalid value for saturation [%v], must be 0-100", *s)
		}
	}
	if b := hsbk.Brightness; b != nil {
		if *b < 0 || *b > 100 {
			return fmt.Errorf("invalid value for brightness [%v], must be 0-100", *b)
		}
	}
	if k := hsbk.Kelvin; k != nil {
		if *k < 1500 || *k > 9000 {
			return fmt.Errorf("invalid value for kelvin [%v], must be 1500-9000", *k)
		}
	}
	return nil
}

func (hsbk *HSBK) IsEmpty() bool {
	return hsbk.Hue == nil && hsbk.Saturation == nil && hsbk.Brightness == nil && hsbk.Kelvin == nil
}

func ValidateActionAndArgs(a Action, hsbk *HSBK) error {
	var hsbkRequired bool
	switch a {
	case ActionPowerOn, ActionPowerOff:
	case ActionPowerSetColor:
		hsbkRequired = true
	case "":
		return fmt.Errorf("action is required")
	default:
		return fmt.Errorf("invalid action: %s", a)
	}

	if err := hsbk.Validate(); err != nil {
		return err
	}

	if hsbkRequired {
		if hsbk == nil || hsbk.IsEmpty() {
			return fmt.Errorf("hsbk must be set for action %s", a)
		}
	}

	return nil
}
