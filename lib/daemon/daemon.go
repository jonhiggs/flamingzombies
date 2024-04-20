package daemon

import (
	"fmt"
	"net"

	"github.com/jonhiggs/flamingzombies/lib/fz"

	log "github.com/sirupsen/logrus"
)

func Listen(c *fz.Config) {
	listener, err := net.Listen("tcp", c.ListenAddress)

	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{
		"file": "lib/daemon/daemon.go",
	}).Info(fmt.Sprintf("Listening for connections on %s", listener.Addr().String()))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection from client: %s", err)
		} else {
			go processClient(conn, c)
		}
	}
}

func processClient(conn net.Conn, c *fz.Config) {
	defer conn.Close()

	fmt.Fprintf(conn, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", "name", "state", "history", "execution_count", "ok_count", "error_count", "last_run", "last_ok")
	for _, t := range c.Tasks {
		fmt.Fprintf(conn, "%s\t%s\t%032b\t%d\t%d\t%d\t%d\t%d\n", t.Name, t.State(), t.History, t.ExecutionCount, t.OKCount, t.ErrorCount, t.LastRun.Unix(), t.LastOk.Unix())
	}
}
