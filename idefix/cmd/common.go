// cmd/common.go
package cmd

import (
	"log/slog"
	"os"
)

// ConfigureLogger sets up a structured logger with appropriate options
func ConfigureLogger(verbose bool, jsonFormat bool) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if verbose {
		opts.Level = slog.LevelDebug
	}

	if jsonFormat {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	logger := slog.New(handler)

	// Set as default logger
	slog.SetDefault(logger)

	return logger
}
