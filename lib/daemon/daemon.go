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
		Name           string    `json:"name"`
		State          string    `json:"state"`
		LastRun        time.Time `json:"last_run"`
		LastOk         time.Time `json:"last_ok"`
		Measurements   []bool    `json:"measurements"`
		ExecutionCount int       `json:"execution_count"`
		OKCount        int       `json:"ok_count"`
		FailCount      int       `json:"fail_count"`
		ErrorCount     int       `json:"error_count"`
	}

	for _, t := range c.Tasks {
		d := taskJsonline{
			Name:           t.Name,
			State:          fmt.Sprintf("%s", t.State()),
			LastRun:        t.LastRun,
			LastOk:         t.LastOk,
			Measurements:   []bool{},
			ExecutionCount: t.ExecutionCount,
			OKCount:        t.OKCount,
			FailCount:      t.FailCount,
			ErrorCount:     t.ErrorCount,
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
