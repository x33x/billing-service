package logger

import (
	"log/slog"
	"os"
	"strings"
)

func Setup(level string) {
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug // -4
	case "warn":
		logLevel = slog.LevelWarn // 4
	case "error":
		logLevel = slog.LevelError // 8
	default:
		logLevel = slog.LevelInfo // 0
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	slog.SetDefault(slog.New(handler))
}
