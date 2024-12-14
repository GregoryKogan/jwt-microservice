package logging

import (
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

func Init() {
	opts := &slog.HandlerOptions{
		Level: getLogLevel(),
	}

	mode := viper.GetString("logging.mode")
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if mode == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	slog.Info("Logging initialized", slog.String("level", opts.Level.Level().String()))
}

func getLogLevel() slog.Level {
	level := viper.GetString("logging.level")
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
