package logger

import (
	"log"
	"log/slog"
	"os"

	"github.com/alessio-palumbo/lifx-force/internal/config"
)

// SetupLogger sets the logging level and output.
// If no logging file is supplied in the config then stdout is used.
func SetupLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler
	level := slog.LevelInfo

	switch *cfg.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	if *cfg.Logging.File != "" {
		f, err := os.OpenFile(*cfg.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}
		handler = slog.NewTextHandler(f, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}
