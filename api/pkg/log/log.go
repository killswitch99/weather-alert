package log

import (
	"log/slog"
	"os"
)

// InitializeLogger sets up the logger
func InitializeLogger() {
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(logHandler))
}