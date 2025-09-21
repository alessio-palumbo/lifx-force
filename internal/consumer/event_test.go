package consumer

import (
	"log/slog"
	"testing"

	"github.com/alessio-palumbo/lifx-force/internal/config"
	"github.com/alessio-palumbo/lifx-force/internal/logger"
	"github.com/alessio-palumbo/lifxlan-go/pkg/device"
	"github.com/alessio-palumbo/lifxlan-go/pkg/messages"
	"github.com/alessio-palumbo/lifxlan-go/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func TestConsumer(t *testing.T) {
	var (
		defaultMs  = 1
		serial0, _ = device.SerialFromHex("d073d5000000")
		serial1, _ = device.SerialFromHex("d073d5000001")
		serial2, _ = device.SerialFromHex("d073d5000002")
		serial3, _ = device.SerialFromHex("d073d5000003")
		label0     = "Desk light"
		label1     = "Night light"
		label2     = "Kids light"
		label3     = "Door light"
		group0     = "Patio"
		group1     = "Bedroom"
		group2     = "Veranda"
		location0  = "Home"
		location1  = "Office"
		devices    = []device.Device{
			{Serial: serial0, Label: label0, Group: group0, Location: location1},
			{Serial: serial1, Label: label1, Group: group1, Location: location0},
			{Serial: serial2, Label: label2, Group: group1, Location: location0},
			{Serial: serial3, Label: label3, Group: group2, Location: location0},
		}
		openHand = config.FingerPattern{1, 1, 1, 1, 1}
	)
	testCases := map[string]struct {
		cfg          *config.Config
		event        *Event
		wantMessages map[device.Serial][]*protocol.Message
	}{
		"gesture event with selector serial": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GestureSwipeLeft,
						Action:   "power_on",
						Selector: config.Selector{Type: config.SelectorTypeSerial, Serial: serial0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Gesture: config.GestureSwipeLeft}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOn()},
			},
		},
		"gesture event with all target": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GestureSwipeLeft,
						Action:   "power_on",
						Selector: config.Selector{Type: config.SelectorTypeAll},
					},
				},
			},
			event: &Event{Hands: []Hand{{Gesture: config.GestureSwipeLeft}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOn()},
				serial1: {messages.SetPowerOn()},
				serial2: {messages.SetPowerOn()},
				serial3: {messages.SetPowerOn()},
			},
		},
		"gesture event with label target": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GestureSwipeLeft,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeLabel, Value: label0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Gesture: config.GestureSwipeLeft}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOff()},
			},
		},
		"gesture event with group target": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GestureSwipeLeft,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeGroup, Value: group1},
					},
				},
			},
			event: &Event{Hands: []Hand{{Gesture: config.GestureSwipeLeft}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial1: {messages.SetPowerOff()},
				serial2: {messages.SetPowerOff()},
			},
		},
		"gesture event with location target": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GestureSwipeLeft,
						Action:   "power_on",
						Selector: config.Selector{Type: config.SelectorTypeLocation, Value: location0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Gesture: config.GestureSwipeLeft}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial1: {messages.SetPowerOn()},
				serial2: {messages.SetPowerOn()},
				serial3: {messages.SetPowerOn()},
			},
		},
		"finger event with selector serial": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Pattern:  &openHand,
						Action:   "power_on",
						Selector: config.Selector{Type: config.SelectorTypeSerial, Serial: serial0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Fingers: config.FingerPattern{1, 1, 1, 1, 1}}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOn()},
			},
		},
		"finger event with all target": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Pattern:  &openHand,
						Action:   "power_on",
						Selector: config.Selector{Type: config.SelectorTypeAll},
					},
				},
			},
			event: &Event{Hands: []Hand{{Fingers: config.FingerPattern{1, 1, 1, 1, 1}}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOn()},
				serial1: {messages.SetPowerOn()},
				serial2: {messages.SetPowerOn()},
				serial3: {messages.SetPowerOn()},
			},
		},
		"finger event with label target": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Pattern:  &openHand,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeLabel, Value: label0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Fingers: config.FingerPattern{1, 1, 1, 1, 1}}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOff()},
			},
		},
		"finger event with group target": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Pattern:  &openHand,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeGroup, Value: group1},
					},
				},
			},
			event: &Event{Hands: []Hand{{Fingers: config.FingerPattern{1, 1, 1, 1, 1}}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial1: {messages.SetPowerOff()},
				serial2: {messages.SetPowerOff()},
			},
		},
		"finger event with location target": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Pattern:  &openHand,
						Action:   "power_on",
						Selector: config.Selector{Type: config.SelectorTypeLocation, Value: location0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Fingers: config.FingerPattern{1, 1, 1, 1, 1}}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial1: {messages.SetPowerOn()},
				serial2: {messages.SetPowerOn()},
				serial3: {messages.SetPowerOn()},
			},
		},
		"ignores fingers when gesture is available": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GestureSwipeLeft,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeLocation, Value: location0},
					},
					{
						Pattern:  &openHand,
						Action:   "power_on",
						Selector: config.Selector{Type: config.SelectorTypeLocation, Value: location0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Fingers: config.FingerPattern{1, 1, 1, 1, 1}, Gesture: config.GestureSwipeLeft}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial1: {messages.SetPowerOff()},
				serial2: {messages.SetPowerOff()},
				serial3: {messages.SetPowerOff()},
			},
		},
		"does nothing with no bindings": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
			},
			event: &Event{Hands: []Hand{{Fingers: config.FingerPattern{1, 1, 1, 1, 1}, Gesture: config.GestureSwipeLeft}}},
		},
		"compound gesture event: expand": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GestureExpand,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeLabel, Value: label0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Label: LeftHandLabel, Gesture: config.GestureSwipeLeft}, {Label: RightHandLabel, Gesture: config.GestureSwipeRight}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOff()},
			},
		},
		"compound gesture event: contract": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GestureContract,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeLabel, Value: label0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Label: LeftHandLabel, Gesture: config.GestureSwipeRight}, {Label: RightHandLabel, Gesture: config.GestureSwipeLeft}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOff()},
			},
		},
		"compound gesture event: push_down": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GesturePushDown,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeLabel, Value: label0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Label: LeftHandLabel, Gesture: config.GestureSwipeDown}, {Label: RightHandLabel, Gesture: config.GestureSwipeDown}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOff()},
			},
		},
		"compound gesture event: pull_up": {
			cfg: &config.Config{
				General: config.General{TransitionMs: defaultMs},
				Bindings: []config.Binding{
					{
						Gesture:  config.GesturePullUp,
						Action:   "power_off",
						Selector: config.Selector{Type: config.SelectorTypeLabel, Value: label0},
					},
				},
			},
			event: &Event{Hands: []Hand{{Label: LeftHandLabel, Gesture: config.GestureSwipeUp}, {Label: RightHandLabel, Gesture: config.GestureSwipeUp}}},
			wantMessages: map[device.Serial][]*protocol.Message{
				serial0: {messages.SetPowerOff()},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctrl := &mockController{devices: devices}
			c := New(tc.cfg, ctrl, logger.NewLogger(slog.LevelInfo, ""))
			c.HandleEvent(tc.event)
			assert.Equal(t, ctrl.messages, tc.wantMessages)
		})
	}
}

type mockController struct {
	devices  []device.Device
	messages map[device.Serial][]*protocol.Message
}

func (m *mockController) Send(serial device.Serial, msg *protocol.Message) error {
	if m.messages == nil {
		m.messages = make(map[device.Serial][]*protocol.Message)
	}
	m.messages[serial] = append(m.messages[serial], msg)
	return nil
}
func (m *mockController) GetDevices() []device.Device {
	return m.devices
}
