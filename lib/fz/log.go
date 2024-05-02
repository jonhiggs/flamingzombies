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

func Fatal(ss ...string) {
	for _, s := range ss {
		fmt.Fprintln(os.Stderr, s)
	}
	os.Exit(1)
}
