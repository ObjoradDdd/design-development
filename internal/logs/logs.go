package logs

import (
	"log/slog"
	"os"
	"strings"
)

func ConfigureSlogLogging(logLevel string) {

	level := slog.LevelDebug

	switch strings.ToLower(logLevel) {
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Log level", "level", level)
}
