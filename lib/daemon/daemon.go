package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/jonhiggs/flamingzombies/lib/fz"
)

func Listen(c *fz.Config) {
	listener, err := net.Listen("tcp", c.ListenAddress)

	if err != nil {
		panic(err)
	}

	fz.Logger.Info(fmt.Sprintf("Listening for connections on %s", listener.Addr().String()))

	for {
		conn, err := listener.Accept()
		if err != nil {
			fz.Logger.Error(fmt.Sprintf("Error accepting connection from client: %s", err))
		} else {
			go processClient(conn, c)
		}
	}
}

func processClient(conn net.Conn, c *fz.Config) {
	defer conn.Close()

	type taskJsonline struct {
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

	for _, t := range c.Tasks {
		d := taskJsonline{
			ErrorCount:       t.ErrorCount,
			ExecutionCount:   t.ExecutionCount,
			FailCount:        t.FailCount,
			LastNotification: t.LastNotification(),
			LastOk:           t.LastOk,
			LastRun:          t.LastRun,
			Measurements:     []bool{},
			Name:             t.Name,
			OKCount:          t.OKCount,
			State:            fmt.Sprintf("%s", t.State()),
		}

		h := t.History
		m := t.HistoryMask
		for h != 0 {
			d.Measurements = append(d.Measurements, (1&h) == 1)
			h = h >> 1
			m = m >> 1
		}

		tjl, err := json.Marshal(d)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(conn, string(tjl))
	}
}
