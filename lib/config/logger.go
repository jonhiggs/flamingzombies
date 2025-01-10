package config

import (
	"fmt"
	"os"

	"log/slog"
)

var Logger *slog.Logger
var logLevel = new(slog.LevelVar)

func StartLogger() {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	Logger = slog.New(h)
}

// Log an error and trigger the error_notifiers
//func Error(traceID string, err error, notifyErrors bool) {
//	Logger.Error(fmt.Sprintf("%s", err))
//	if notifyErrors {
//		for _, errN := range cfg.Defaults.ErrorNotifierNames {
//			ErrorNotifyCh <- ErrorNotification{
//				Notifier: cfg.GetNotifierByName(errN),
//				Error:    err,
//				TraceID:  traceID,
//			}
//		}
//	}
//}

//func Fatal(ss ...string) {
//	for _, s := range ss {
//		fmt.Fprintln(os.Stderr, s)
//	}
//	os.Exit(1)
//}

func SetLogLevel(l string) error {
	switch l {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		return fmt.Errorf("invalid log level: %s", l)
	}

	return nil
}
