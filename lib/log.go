package fz

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func StartLogger(l string) {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	if config.LogFile == "stdout" || config.LogFile == "-" {
		log.SetOutput(os.Stdout)
	} else if config.LogFile == "stderr" {
		log.SetOutput(os.Stderr)
	} else {
		f, err := os.OpenFile(config.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
	}

	switch l {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}
}
