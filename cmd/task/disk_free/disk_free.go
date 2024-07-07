package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/cli"
	"golang.org/x/sys/unix"
	"nullprogram.com/x/optparse"
)

const (
	APP             = "disk_free"
	TIMEOUT_DEFAULT = 5 * time.Second
	TIMEOUT_MINIMUM = 2 * time.Second
	VERSION         = "v0.0.0"
)

var (
	path    string
	bytes   = make(chan int64, 1)
	timeout time.Duration
)

func init() {
	// set defaults
	timeout = cli.GetTimeout(0, TIMEOUT_DEFAULT, TIMEOUT_MINIMUM)

	options := []optparse.Option{
		{"debug", 'd', optparse.KindNone},
		{"help", 'h', optparse.KindNone},
		{"timeout", 't', optparse.KindRequired},
		{"version", 'V', optparse.KindNone},
	}

	results, rest, err := optparse.Parse(options, os.Args)
	if err != nil {
		cli.Error(err)
	}

	for _, result := range results {
		switch result.Long {
		case "debug":
			cli.Debug = true
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
			cli.Version(APP, VERSION)
		}
	}

	if len(rest) != 1 {
		cli.Error("path was not provided")
	}

	path = rest[0]
}

func disk_free(dir string) {
	var stat unix.Statfs_t
	unix.Statfs(dir, &stat)

	// Available blocks * size per block = available space in bytes, then
	// converted to KB
	free := (stat.Bavail * uint64(stat.Bsize)) / 1024
	bytes <- int64(free)
}

func main() {
	go disk_free(path)

	for {
		select {
		case b := <-bytes:
			if b == 0 {
				cli.Error(fmt.Sprint("could not measure free space on ", path))
				os.Exit(3)
			}

			cli.StatsdValue(b)
			os.Exit(0)
		case <-time.After(timeout):
			// if no packets were received by the timeout, then the host is down.
			fmt.Printf("no packets received within timeout (%v)\n", timeout)
			os.Exit(1)
		}
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Printf("  %s [OPTIONS] PATH\n", APP)
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -d, --debug    Print debugging info")
	fmt.Println("  -h, --help     This help")
	fmt.Println("  -t, --timeout  Seconds to wait for a response")
	fmt.Println("  -V, --version  Version")
	os.Exit(0)
}
