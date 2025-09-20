package runtime

import (
	"testing"

	"github.com/alessio-palumbo/lifx-force/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestArgsFromConfig(t *testing.T) {
	cfg0 := &config.Config{
		Tracking: config.Tracking{
			FrameSkip:        1,
			BufferSize:       5,
			GestureThreshold: 0.1,
		},
	}
	want := []string{
		"--frame-skip", "1",
		"--buffer-size", "5",
		"--gesture-threshold", "0.1",
	}
	assert.Equal(t, want, ArgsFromConfig(cfg0))
}
