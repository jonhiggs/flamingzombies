package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"nullprogram.com/x/optparse"
)

const (
	DEFAULT_TIMEOUT_SECONDS = 1
	VERSION                 = "v0.0.0"
)

var (
	retries int = 3
	address string
	timeout time.Duration
	pkts    = make(chan *probing.Packet, 1)
)

func init() {
	if os.Getenv("TIMEOUT") == "" {
		timeout = DEFAULT_TIMEOUT_SECONDS * time.Second
	} else {
		t, err := strconv.Atoi(os.Getenv("TIMEOUT"))
		if err != nil {
			log.Fatal(err)
		}
		timeout = time.Duration(t) * time.Second
	}

	options := []optparse.Option{
		{"retries", 'r', optparse.KindRequired},
		{"help", 'h', optparse.KindNone},
		{"version", 'V', optparse.KindNone},
	}

	results, rest, err := optparse.Parse(options, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		switch result.Long {
		case "retries":
			var err error
			retries, err = strconv.Atoi(result.Optarg)

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
		log.Fatal("No hosts were provided")
	}

	address = rest[0]
}

func ping() {
	pinger, err := probing.NewPinger(address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pinger.OnRecv = func(pkt *probing.Packet) {
		pkts <- pkt
		pinger.Stop()
	}

	err = pinger.Run()
	if err != nil {
		panic(err)
	}

}

func main() {
	go ping()

	for {
		select {
		case p := <-pkts:
			fmt.Printf("%v\n", p.Rtt)
			os.Exit(0)
		case <-time.After(timeout):
			fmt.Println("out of time :(")
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  ping [OPTIONS] HOST")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -r, --retries <number>  Attempts for ICMP response")
	fmt.Println("  -h, --help              This help")
	fmt.Println("  -V, --version           Version")
	os.Exit(0)
}
