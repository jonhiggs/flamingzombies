package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/cli"
	probing "github.com/prometheus-community/pro-bing"
	"nullprogram.com/x/optparse"
)

const (
	APP             = "ping"
	TIMEOUT_DEFAULT = 5 * time.Second
	TIMEOUT_MINIMUM = 2 * time.Second
	VERSION         = "v0.0.0"

	PACKETS = 5
)

var (
	address string
	timeout time.Duration
	pkts    = make(chan *probing.Packet, 1)
)

func init() {
	// set defaults
	timeout = cli.GetTimeout(0, TIMEOUT_DEFAULT, TIMEOUT_MINIMUM)

	options := []optparse.Option{
		{"timeout", 't', optparse.KindRequired},
		{"help", 'h', optparse.KindNone},
		{"version", 'V', optparse.KindNone},
	}

	results, rest, err := optparse.Parse(options, os.Args)
	if err != nil {
		cli.Error(err)
	}

	for _, result := range results {
		switch result.Long {
		case "timeout":
			t, err := strconv.Atoi(result.Optarg)
			if err != nil {
				cli.Error(err)
			}

			td := time.Duration(t) * time.Second
			timeout = cli.GetTimeout(td, TIMEOUT_DEFAULT, TIMEOUT_MINIMUM)

		case "help":
			usage()
		case "version":
			cli.Version("task/ping", VERSION)
		}
	}

	if len(rest) != 1 {
		cli.Error("No host targets were provided")
	}

	address = rest[0]
}

func ping() {
	pinger, err := probing.NewPinger(address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pinger.Count = 5

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
			// if no packets were received by the timeout, then the host is down.
			fmt.Printf("No packets received within timeout (%v)\n", timeout)
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Printf("  %s [OPTIONS] HOST\n", APP)
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -t, --timeout  Seconds to wait for a response")
	fmt.Println("  -h, --help     This help")
	fmt.Println("  -V, --version  Version")
	os.Exit(0)
}
