package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/daemon"
	"github.com/jonhiggs/flamingzombies/lib/fz"
	"nullprogram.com/x/optparse"
)

const VERSION = "v0.0.17"

var config fz.Config
var configtestMode = false

func init() {
	if os.Getenv("FZ_CONFIG_FILE") == "" {
		os.Setenv("FZ_CONFIG_FILE", "/etc/flamingzombies.toml")
	}

	if os.Getenv("FZ_DIRECTORY") == "" {
		os.Setenv("FZ_DIRECTORY", "/usr/libexec/flamingzombies")
	}

	if os.Getenv("FZ_LOG_LEVEL") == "" {
		os.Setenv("FZ_LOG_LEVEL", "info")
	}

	if os.Getenv("FZ_LOG_FILE") == "" {
		os.Setenv("FZ_LOG_FILE", "stdout")
	}

	options := []optparse.Option{
		{"config", 'c', optparse.KindRequired},
		{"configtest", 'n', optparse.KindNone},
		{"directory", 'C', optparse.KindRequired},
		{"help", 'h', optparse.KindNone},
		{"loglevel", 'l', optparse.KindRequired},
		{"version", 'V', optparse.KindNone},
	}

	results, _, err := optparse.Parse(options, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		switch result.Long {
		case "config":
			os.Setenv("FZ_CONFIG_FILE", result.Optarg)
		case "configtest":
			configtestMode = true
		case "loglevel":
			os.Setenv("FZ_LOG_LEVEL", result.Optarg)
		case "directory":
			os.Setenv("FZ_DIRECTORY", result.Optarg)
		case "help":
			usage()
			return
		case "version":
			fmt.Printf("flamingzombies %s\n", VERSION)
			os.Exit(0)
		}
	}

	config = fz.ReadConfig()

	if os.Getenv("FZ_LOG_LEVEL") != "" {
		fz.StartLogger(os.Getenv("FZ_LOG_LEVEL"))
	} else {
		fz.StartLogger(config.LogLevel)
	}

	if configtestMode {
		fmt.Println("The configuration is valid")
		os.Exit(0)
	}

	fz.ProcessNotifications()

	if config.Listen() {
		go daemon.Listen(&config)
	}
}

func main() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			for i, t := range config.Tasks {
				if t.Ready(ts) {
					go config.Tasks[i].Run()
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
	fmt.Println("  -c, --config <file>     Configuration file")
	fmt.Println("  -C, --directory <path>  Change to directory")
	fmt.Println("  -h, --help              This help")
	fmt.Println("  -l, --loglevel <level>  Override the log level")
	fmt.Println("  -n, --configtest        Test validity of the configuration")
	fmt.Println("  -V, --version           Version")
	os.Exit(0)
}
