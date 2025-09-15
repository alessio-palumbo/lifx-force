package consumer

import (
	"fmt"

	"github.com/alessio-palumbo/lifx-force/internal/config"
	"github.com/alessio-palumbo/lifxlan-go/pkg/device"
	"github.com/alessio-palumbo/lifxlan-go/pkg/protocol"
)

type Hand struct {
	Label   string  `json:"label"`
	Fingers []int   `json:"fingers"`
	Gesture *string `json:"gesture"`
}

type Event struct {
	Hands []Hand `json:"hands"`
}

type lanController interface {
	Send(serial device.Serial, msg *protocol.Message) error
	GetDevices() []device.Device
}

type Consumer struct {
	ctrl lanController
	cfg  *config.Config
}

func New(cfg *config.Config, ctrl lanController) *Consumer {
	return &Consumer{cfg: cfg, ctrl: ctrl}
}

func (*Consumer) HandleEvent(event *Event) {
	// Placeholder
	fmt.Println(event)
}
