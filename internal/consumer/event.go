package consumer

import (
	"log/slog"
	"time"

	"github.com/alessio-palumbo/lifx-force/internal/config"
	"github.com/alessio-palumbo/lifxlan-go/pkg/device"
	"github.com/alessio-palumbo/lifxlan-go/pkg/messages"
	"github.com/alessio-palumbo/lifxlan-go/pkg/protocol"
	"github.com/alessio-palumbo/lifxprotocol-go/gen/protocol/enums"
)

var compoundGestures = map[config.Gesture]func(map[label]Hand) bool{
	config.GestureExpand: func(hs map[label]Hand) bool {
		return matchGesture(hs[LeftHandLabel], "swipe_left") &&
			matchGesture(hs[RightHandLabel], "swipe_right")
	},
	config.GestureContract: func(hs map[label]Hand) bool {
		return matchGesture(hs[LeftHandLabel], "swipe_right") &&
			matchGesture(hs[RightHandLabel], "swipe_left")
	},
	config.GesturePushDown: func(hs map[label]Hand) bool {
		return matchGesture(hs[LeftHandLabel], "swipe_down") &&
			matchGesture(hs[RightHandLabel], "swipe_down")
	},
	config.GesturePullUp: func(hs map[label]Hand) bool {
		return matchGesture(hs[LeftHandLabel], "swipe_up") &&
			matchGesture(hs[RightHandLabel], "swipe_up")
	},
}

type label string

const (
	LeftHandLabel  = "left"
	RightHandLabel = "right"
)

type Hand struct {
	Label   label                `json:"label"`
	Fingers config.FingerPattern `json:"fingers"`
	Gesture config.Gesture       `json:"gesture,omitempty"`
}

type Event struct {
	Hands []Hand `json:"hands"`
}

type lanController interface {
	Send(serial device.Serial, msg *protocol.Message) error
	GetDevices() []device.Device
}

type sendFunc func(ctrl lanController) error

type Consumer struct {
	ctrl            lanController
	cfg             *config.Config
	logger          *slog.Logger
	fingerBindings  map[config.FingerPattern]sendFunc
	gestureBindings map[config.Gesture]sendFunc
}

func New(cfg *config.Config, ctrl lanController, logger *slog.Logger) *Consumer {
	gb, fb := initBindings(cfg, logger, ctrl.GetDevices())
	return &Consumer{
		cfg:             cfg,
		ctrl:            ctrl,
		logger:          logger,
		gestureBindings: gb,
		fingerBindings:  fb,
	}
}

func (c *Consumer) HandleEvent(event *Event) {
	c.logger.Debug("processing event", slog.Any("event", event))

	hs := make(map[label]Hand, len(event.Hands))
	for _, h := range event.Hands {
		hs[h.Label] = h
	}

	// Try compound gestures
	for g, match := range compoundGestures {
		if match(hs) {
			if f, ok := c.gestureBindings[g]; ok {
				c.logger.Debug("actioned compound gesture", slog.Any("gesture", g))
				f(c.ctrl)
				return
			}
			c.logger.Debug("unhandled compound gesture", slog.Any("gesture", g))
			break
		}
	}

	// Fallback: single-hand gestures
	for _, h := range event.Hands {
		if h.Gesture != "" && c.gestureBindings != nil {
			if f, ok := c.gestureBindings[h.Gesture]; ok {
				c.logger.Debug("actioned gesture", slog.Any("gesture", h.Gesture))
				f(c.ctrl)
				// Skip finger binding when gesture is available.
				continue
			}
			c.logger.Warn("unhandled gesture", slog.Any("hand", h.Label), slog.Any("gesture", h.Gesture))
		}

		if c.fingerBindings != nil {
			if f, ok := c.fingerBindings[h.Fingers]; ok {
				f(c.ctrl)
				continue
			}
			c.logger.Warn("unhandled finger binding", slog.Any("hand", h.Label), slog.Any("fingers", h.Fingers))
		}
	}
}

func initBindings(cfg *config.Config, logger *slog.Logger, devices []device.Device) (map[config.Gesture]sendFunc, map[config.FingerPattern]sendFunc) {
	gb := make(map[config.Gesture]sendFunc)
	fb := make(map[config.FingerPattern]sendFunc)
	for _, b := range cfg.Bindings {
		if b.Gesture != "" {
			if f := bindingSendFunc(cfg, devices, b.Action, b.HSBK, b.Selector); f != nil {
				gb[b.Gesture] = f
				logger.Debug("registered gesture binding", slog.Any("gesture", b.Gesture))
			}
			continue
		}
		if f := bindingSendFunc(cfg, devices, b.Action, b.HSBK, b.Selector); f != nil {
			fb[*b.Pattern] = f
			logger.Debug("registered finger binding", slog.Any("fingers", b.Pattern))
		}
	}
	return gb, fb
}

func bindingSendFunc(cfg *config.Config, devices []device.Device, action config.Action, hsbk *config.HSBK, selector config.Selector) sendFunc {
	var msg *protocol.Message
	switch action {
	case config.ActionPowerOn:
		msg = messages.SetPowerOn()
	case config.ActionPowerOff:
		msg = messages.SetPowerOff()
	case config.ActionPowerSetColor:
		msg = messages.SetColor(
			hsbk.Hue, hsbk.Saturation, hsbk.Brightness, hsbk.Kelvin,
			time.Duration(cfg.General.TransitionMs)*time.Millisecond, enums.LightWaveformLIGHTWAVEFORMSAW,
		)
	default:
		return nil
	}

	var serials []device.Serial
	switch selector.Type {
	case config.SelectorTypeAll:
		serials = targetForCondition(devices, func(d *device.Device) bool { return true })
	case config.SelectorTypeLabel:
		serials = targetForCondition(devices, func(d *device.Device) bool { return d.Label == selector.Value })
	case config.SelectorTypeGroup:
		serials = targetForCondition(devices, func(d *device.Device) bool { return d.Group == selector.Value })
	case config.SelectorTypeLocation:
		serials = targetForCondition(devices, func(d *device.Device) bool { return d.Location == selector.Value })
	case config.SelectorTypeSerial:
		serials = []device.Serial{selector.Serial}
	default:
		return nil
	}

	return func(ctrl lanController) error { return sendMultiple(ctrl, serials, msg) }
}

func targetForCondition(devices []device.Device, cond func(d *device.Device) bool) []device.Serial {
	var targets []device.Serial
	for _, d := range devices {
		if cond(&d) {
			targets = append(targets, d.Serial)
		}
	}
	return targets
}

func sendMultiple(ctrl lanController, serials []device.Serial, msg *protocol.Message) error {
	for _, s := range serials {
		if err := ctrl.Send(s, msg); err != nil {
			return err
		}
	}
	return nil
}

func matchGesture(h Hand, g config.Gesture) bool {
	return h.Gesture != "" && h.Gesture == g
}
