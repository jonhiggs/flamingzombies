package daemon

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	listen = flag.Bool("l", false, "Listen")
	host   = flag.String("h", "localhost", "Host")
	port   = flag.Int("p", 5891, "Port")
)

func Listen() {
	// compose server address from host and port
	addr := fmt.Sprintf("%s:%d", *host, *port)
	// launch TCP server
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		// if we can't launch server for some reason,
		// we can't do anything about it, just panic!
		panic(err)
	}

	log.Printf("Listening for connections on %s", listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection from client: %s", err)
		} else {
			go processClient(conn)
		}
	}
}

func processClient(conn net.Conn) error {
	defer conn.Close()

	fmt.Fprintf(conn, "# ")

	buff := make([]byte, 1024)
	c := bufio.NewReader(conn)

	for {
		size, err := c.ReadByte()
		if err != nil {
			return err
		}

		// read the full message, or return an error
		_, err = io.ReadFull(c, buff[:int(size)])
		if err != nil {
			return err

		}
	}
}
