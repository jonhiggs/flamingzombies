package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/fz"
	"nullprogram.com/x/optparse"
)

const VERSION = "v0.0.21"

var configTest = false
var configFile = "/etc/flamingzombies.toml"
var dir = "/usr/libexec/flamingzombies"
var logLevel = "info"
var cfg *fz.Config

func init() {
	options := []optparse.Option{
		{"config", 'c', optparse.KindRequired},
		{"configtest", 'n', optparse.KindNone},
		{"directory", 'C', optparse.KindRequired},
		{"help", 'h', optparse.KindNone},
		{"loglevel", 'l', optparse.KindRequired},
		{"pidfile", 'p', optparse.KindRequired},
		{"statsd-host", ' ', optparse.KindRequired},
		{"statsd-prefix", ' ', optparse.KindRequired},
		{"version", 'V', optparse.KindNone},
	}

	results, _, err := optparse.Parse(options, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		switch result.Long {
		case "config":
			configFile = result.Optarg
		case "configtest":
			configTest = true
		case "loglevel":
			logLevel = result.Optarg
		case "directory":
			dir = result.Optarg
		case "pidfile":
			err := os.WriteFile(result.Optarg, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "help":
			usage()
			return
		case "version":
			fmt.Printf("fz %s\n", VERSION)
			os.Exit(0)
		}
	}

	if os.Getenv("FZ_CONFIG_FILE") != "" {
		configFile = os.Getenv("FZ_CONFIG_FILE")
	}

	cfg = fz.ReadConfig(configFile)

	// working directory
	if os.Getenv("FZ_DIRECTORY") == "" {
		cfg.Directory = dir
	} else {
		cfg.Directory = os.Getenv("FZ_DIRECTORY")
	}

	// logging
	if os.Getenv("FZ_LOG_LEVEL") == "" {
		cfg.LogLevel = logLevel
	} else {
		cfg.LogLevel = os.Getenv("FZ_LOG_LEVEL")
	}
	fz.StartLogger(cfg.LogLevel)

	// validation
	if err = cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	if configTest {
		// break out if we're in config test mode.
		fmt.Println("The configuration is valid")
		os.Exit(0)
	}

	fz.ProcessNotifications()
}

func main() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			for i, t := range cfg.Tasks {
				if t.Ready(ts) {
					go cfg.Tasks[i].Run()
				}
			}
		}
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  fz [OPTIONS]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -c, --config <file>            Configuration file")
	fmt.Println("  -n, --configtest               Test validity of the configuration")
	fmt.Println("  -C, --directory <path>         Change to directory")
	fmt.Println("  -h, --help                     This help")
	fmt.Println("  -l, --loglevel <level>         Override the log level")
	fmt.Println("  -p, --pidfile                  The pidfile to write")
	fmt.Println("  -V, --version                  Version")
	os.Exit(0)
}
