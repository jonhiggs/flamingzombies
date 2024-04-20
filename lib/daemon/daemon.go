package daemon

import (
	"flag"
	"fmt"
	"net"

	"git.altos/flamingzombies/lib/fz"

	log "github.com/sirupsen/logrus"
)

var (
	listen = flag.Bool("l", false, "Listen")
	host   = flag.String("h", "localhost", "Host")
	port   = flag.Int("p", 5891, "Port")
)

func Listen(c *fz.Config) {
	addr := fmt.Sprintf("%s:%d", *host, *port)
	listener, err := net.Listen("tcp", addr)

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

	for _, t := range c.Tasks {
		fmt.Fprintf(conn, "%s\t%s\t%032b\t%d\t%d\n", t.Name, t.State(), t.History, t.LastRun.Unix(), t.LastOk.Unix())
	}

}
