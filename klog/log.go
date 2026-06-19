package klog

import (
	"log/slog"
	"os"

	"git.mypierian.com/borghives/kosmos-go/ether"
)

func Ignite() *slog.Logger {
	logLevel := ether.UniversalConstants.Collapse().DebugLevel
	logger := NewLogger(logLevel)
	slog.SetDefault(logger)
	return logger
}

func NewLogger(level int) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.Level(level),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.MessageKey:
				a.Key = "message"
			case slog.LevelKey:
				a.Key = "severity"
				// GCP expects "WARNING" instead of the slog default "WARN"
				if level, ok := a.Value.Any().(slog.Level); ok && level == slog.LevelWarn {
					a.Value = slog.StringValue("WARNING")
				}
			}
			return a
		},
	}

	// Outputting to os.Stdout is perfectly fine; GCP agents will pick it up automatically.
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func Err(err error) slog.Attr {
	return slog.String("error", err.Error())
}
