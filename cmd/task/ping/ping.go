package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	probing "github.com/prometheus-community/pro-bing"
	"nullprogram.com/x/optparse"
)

const (
	DEFAULT_TIMEOUT_SECONDS = "5"
	VERSION                 = "v0.0.0"
)

var (
	count   int = 3
	address string
)

func init() {
	if os.Getenv("TIMEOUT") == "" {
		os.Setenv("TIMEOUT", DEFAULT_TIMEOUT_SECONDS)
	}

	options := []optparse.Option{
		{"count", 'c', optparse.KindRequired},
		{"help", 'h', optparse.KindNone},
		{"version", 'V', optparse.KindNone},
	}

	results, rest, err := optparse.Parse(options, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		switch result.Long {
		case "count":
			var err error
			count, err = strconv.Atoi(result.Optarg)

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

	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "Error: No hosts were provided")
		os.Exit(1)
	}

	address = rest[0]
}

func main() {
	pinger, err := probing.NewPinger(address)
	if err != nil {
		panic(err)
	}
	pinger.Count = count
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		panic(err)
	}
	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats
	fmt.Println(stats)
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  ping [OPTIONS] HOST")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -c, --count <number>           Number of ICMP packets to send")
	fmt.Println("  -h, --help                     This help")
	fmt.Println("  -V, --version                  Version")
	os.Exit(0)
}
