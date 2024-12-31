package fz

import (
	"fmt"
	"log/slog"
	"os"
)

var Logger *slog.Logger
var loggerLevel = new(slog.LevelVar)

func StartLogger(l string) {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLevel})
	Logger = slog.New(h)

	switch l {
	case "debug":
		loggerLevel.Set(slog.LevelDebug)
	case "info":
		loggerLevel.Set(slog.LevelInfo)
	case "warn":
		loggerLevel.Set(slog.LevelWarn)
	case "error":
		loggerLevel.Set(slog.LevelError)
	default:
		panic(fmt.Sprintf("Invalid log level: %s", l))
	}
}

// Log an error and trigger the error_notifiers
func Error(traceID string, err error, notifyErrors bool) {
	Logger.Error(fmt.Sprintf("%s", err))
	if notifyErrors {
		for _, errN := range cfg.Defaults.ErrorNotifierNames {
			ErrorNotifyCh <- ErrorNotification{
				Notifier: cfg.GetNotifierByName(errN),
				Error:    err,
				TraceID:  traceID,
			}
		}
	}
}

func Fatal(ss ...string) {
	for _, s := range ss {
		fmt.Fprintln(os.Stderr, s)
	}
	os.Exit(1)
}
