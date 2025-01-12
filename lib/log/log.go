package log

import (
	"fmt"
	"os"

	"log/slog"
)

var logger *slog.Logger
var logLevel = new(slog.LevelVar)

func init() {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger = slog.New(h)
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

func Fatal(msgs ...string) {
	for _, m := range msgs {
		fmt.Fprintln(os.Stderr, m)
	}
	os.Exit(1)
}

// Adjust the log level
func SetLevel(l string) error {
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

// A wrapper for slog.Debug()
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// A wrapper for slog.Info()
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// A wrapper for slog.Info()
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// A wrapper for slog.Error()
func Error(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Log a system error
// XXX(jh) 20250112: not sure if this will actually be needed...
func SystemError(err error) {
	Error(fmt.Sprint(err))
}
