package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"nullprogram.com/x/optparse"
)

const VERSION = "v0.0.21"

type taskData struct {
	ErrorCount       int       `json:"error_count"`
	ExecutionCount   int       `json:"execution_count"`
	FailCount        int       `json:"fail_count"`
	LastNotification time.Time `json:"last_notification"`
	LastOk           time.Time `json:"last_ok"`
	LastRun          time.Time `json:"last_run"`
	Measurements     []bool    `json:"measurements"`
	Name             string    `json:"name"`
	OKCount          int       `json:"ok_count"`
	State            string    `json:"state"`
}

var tasks []taskData
var cmd = "list"
var subcmds = []string{}

func init() {
	if os.Getenv("FZ_LISTEN") == "" {
		os.Setenv("FZ_LISTEN", "127.0.0.1:5891")
	}

	options := []optparse.Option{
		{"help", 'h', optparse.KindNone},
		{"host", 'H', optparse.KindRequired},
		{"version", 'V', optparse.KindNone},
	}

	results, _, err := optparse.Parse(options, os.Args)
	if err != nil {
		log.Fatal(err)
	}

	optCount := 0
	for _, result := range results {
		switch result.Long {
		case "help":
			usage()
			return
		case "host":
			os.Setenv("FZ_LISTEN", result.Optarg)
			optCount += 2
		case "version":
			fmt.Printf("fzctl %s\n", VERSION)
			os.Exit(0)
		}
	}

	if len(os.Args) > optCount+1 {
		cmd = os.Args[optCount+1]
	}

	if len(os.Args) > optCount+2 {
		subcmds = os.Args[optCount+2:]
	}
}

func main() {
	switch cmd {
	case "list":
		list()
	case "show":
		show()
	case "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func list() {
	for _, s := range subcmds {
		if s != "ok" && s != "fail" && s != "unknown" {
			fmt.Fprintf(os.Stderr, "Unsupported subcommand: %s\n", s)
			fmt.Fprintln(os.Stderr, "  allowed values are 'ok', 'fail', or 'unknown'.")
			os.Exit(1)
		}
	}

	fetchTasks()

	if len(subcmds) == 0 {
		subcmds = []string{"ok", "fail", "unknown"}
	}
	for _, s := range subcmds {
		for _, t := range tasks {
			if t.State == s {
				fmt.Printf("%-40s\t%s\n", t.Name, t.State)
			}
		}
	}
}

func show() {
	if len(subcmds) == 0 {
		fmt.Fprintln(os.Stderr, "No task was provided.")
		os.Exit(1)
	}

	fetchTasks()

	for i, s := range subcmds {
		for _, t := range tasks {
			if t.Name == s {
				if i != 0 {
					fmt.Println("")
				}

				fmt.Printf("%-20s%s\n", "name:", t.Name)
				fmt.Printf("%-20s%s\n", "state:", t.State)
				fmt.Printf("%-20s%s ago\n", "last execution:", time.Now().Sub(t.LastRun))
				fmt.Printf("%-20s%s ago\n", "last notification:", time.Now().Sub(t.LastNotification))
				fmt.Printf("%-20s%s ago\n", "last success:", time.Now().Sub(t.LastOk))
				fmt.Printf("%-20s%d\n", "executions:", t.ExecutionCount)
				fmt.Printf("%-20s%d\n", "failures:", t.FailCount)
				fmt.Printf("%-20s%d\n", "errors:", t.ErrorCount)
			}
		}
	}
}

func fetchTasks() {
	data, err := readFromNetwork(fmt.Sprintf("%s", os.Getenv("FZ_LISTEN")))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		var t taskData
		err = json.Unmarshal(scanner.Bytes(), &t)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}

		tasks = append(tasks, t)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func readFromNetwork(addr string) ([]byte, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return []byte{}, err
	}
	defer conn.Close()

	buf := make([]byte, 1)
	var data []byte

	for {
		_, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return []byte{}, err
		}

		data = append(data[:], buf[:]...)
	}

	return data, nil

	//fmt.Println("Read", len(data), "bytes from the connection:", string(data[:len(data)]))
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  fzctl [OPTIONS] <command> [subcommand]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -H, --host <ip:port>        Host to connect to")
	fmt.Println("  -V, --version               Version")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  help                        This help")
	fmt.Println("  list [ok,fail,unknown]      List the running tasks")
	fmt.Println("  show <task>                 Show the details of a task")
	os.Exit(0)
}
