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

type Hand struct {
	Label   string               `json:"label"`
	Fingers config.FingerPattern `json:"fingers"`
	Gesture *string              `json:"gesture"`
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
	return &Consumer{
		cfg:             cfg,
		ctrl:            ctrl,
		logger:          logger,
		fingerBindings:  initFingerBindings(cfg, ctrl.GetDevices()),
		gestureBindings: initGestureBindings(cfg, ctrl.GetDevices()),
	}
}

func (c *Consumer) HandleEvent(event *Event) {
	c.logger.Debug("processing event", slog.Any("event", event))
	for _, h := range event.Hands {
		if h.Gesture != nil && c.gestureBindings != nil {
			if f, ok := c.gestureBindings[config.Gesture(*h.Gesture)]; ok {
				f(c.ctrl)
				// Skip finger binding when gesture is available.
				continue
			}
			c.logger.Warn("unhandled gesture", slog.String("gesture", *h.Gesture))
		}

		if c.fingerBindings != nil {
			if f, ok := c.fingerBindings[h.Fingers]; ok {
				f(c.ctrl)
				continue
			}
			c.logger.Warn("unhandled finger binding", slog.Any("fingers", h.Fingers))
		}
	}
}

func initFingerBindings(cfg *config.Config, devices []device.Device) map[config.FingerPattern]sendFunc {
	m := make(map[config.FingerPattern]sendFunc)
	for _, b := range cfg.FingerBindings {
		if f := bindingSendFunc(cfg, devices, b.Action, b.HSBK, b.Selector); f != nil {
			m[b.Pattern] = f
		}
	}
	return m
}

func initGestureBindings(cfg *config.Config, devices []device.Device) map[config.Gesture]sendFunc {
	m := make(map[config.Gesture]sendFunc)
	for _, b := range cfg.GestureBindings {
		switch b.Gesture {
		case config.GestureSwipeLeft, config.GestureSwipeRight:
		default:
			continue
		}

		if f := bindingSendFunc(cfg, devices, b.Action, b.HSBK, b.Selector); f != nil {
			m[b.Gesture] = f
		}
	}
	return m
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
