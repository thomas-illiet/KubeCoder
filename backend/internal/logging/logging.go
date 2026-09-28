package logging

import (
	"io"
	"log/slog"

	"github.com/thomas-illiet/KubeCoder/backend/internal/config"
)

// New creates a structured logger from the application configuration.
func New(cfg config.LogConfig, output io.Writer) (*slog.Logger, error) {
	level, err := config.ParseLogLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	opts := &slog.HandlerOptions{Level: level}
	if cfg.Format == "text" {
		return slog.New(slog.NewTextHandler(output, opts)), nil
	}
	return slog.New(slog.NewJSONHandler(output, opts)), nil
}
