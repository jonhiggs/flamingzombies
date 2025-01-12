package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/jonhiggs/flamingzombies/lib/config"
	"github.com/jonhiggs/flamingzombies/lib/core"
	"github.com/jonhiggs/flamingzombies/lib/log"
	"nullprogram.com/x/optparse"
)

func init() {
	var configFile = "/etc/flamingzombies.toml"
	var configTest = false
	var directory, logLevel string

	options := []optparse.Option{
		{"config", 'f', optparse.KindRequired},
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
		log.Fatal(fmt.Sprint(err))
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
			directory = result.Optarg
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
			fmt.Printf("fz %s\n", config.VERSION)
			os.Exit(0)
		}
	}

	if os.Getenv("FZ_CONFIG_FILE") != "" {
		configFile = os.Getenv("FZ_CONFIG_FILE")
	}

	if os.Getenv("FZ_DIRECTORY") != "" {
		if err := config.SetDirectory(os.Getenv("FZ_DIRECTORY")); err != nil {
			log.Fatal(fmt.Sprint(err))
		}
	}

	if os.Getenv("FZ_LOG_LEVEL") != "" {
		logLevel = os.Getenv("FZ_LOG_LEVEL")
	}

	cfgFh, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	if err := config.Load(cfgFh); err != nil {
		panic(err)
	}

	// when there is an override to the log level
	if len(logLevel) > 0 {
		if err := log.SetLevel(logLevel); err != nil {
			// TODO(jh) 20250111: convert to an fatal error
			panic(err)
		}
	} else {
		if err := log.SetLevel(config.LogLevel); err != nil {
			// TODO(jh) 20250111: convert to an fatal error
			panic(err)
		}
	}

	// when there is an override to the directory
	if len(directory) > 0 {
		if err := config.SetDirectory(directory); err != nil {
			log.Fatal(fmt.Sprint(err))
		}
	}

	// validation
	//if err = config.Validate(); err != nil {
	//	log.Fatal(err)
	//}
	if configTest {
		// break out if we're in config test mode.
		fmt.Println("The configuration is valid")
		os.Exit(0)
	}
}

func main() {
	var wg sync.WaitGroup
	go core.ScheduleTasks()
	wg.Add(1)
	go core.ProcessTasks()
	wg.Add(1)

	wg.Wait()
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  fz [OPTIONS]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -C, --directory <path>         Change to directory")
	fmt.Println("  -f, --config <file>            Configuration file")
	fmt.Println("  -h, --help                     This help")
	fmt.Println("  -l, --loglevel <level>         Override the log level")
	fmt.Println("  -n, --configtest               Test validity of the configuration")
	fmt.Println("  -p, --pidfile                  The pidfile to write")
	fmt.Println("  -V, --version                  Version")
	os.Exit(0)
}
