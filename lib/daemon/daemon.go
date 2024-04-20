package daemon

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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

func processClient(conn net.Conn) {
	_, err := io.Copy(os.Stdout, conn)
	if err != nil {
		fmt.Println(err)
	}
	conn.Close()
}
