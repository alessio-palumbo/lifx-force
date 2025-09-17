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
	level := slog.LevelInfo

	switch cfg.Logging.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	return NewLogger(level, cfg.Logging.File)
}

func NewLogger(level slog.Level, logFile string) *slog.Logger {
	var handler slog.Handler
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}
		handler = slog.NewTextHandler(f, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}
