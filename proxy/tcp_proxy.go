package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
)

type TcpProxy struct {
	Port       string
	TargetAddr string
}

func (p TcpProxy) handleConnection(conn net.Conn) {
	connAddress := fmt.Sprintf("%v:%v", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	log.Printf("Received connection %s\n", connAddress)

	targetConn, err := net.Dial("tcp", p.TargetAddr)
	if err != nil {
		conn.Close()
		log.Fatalf("Can't connect to %v: %v", p.TargetAddr, err)
	}

	go io.Copy(targetConn, conn)
	go io.Copy(conn, targetConn)
}

func (p TcpProxy) Start() {
	ln, err := net.Listen("tcp", p.Port)
	defer ln.Close()

	if err != nil {
		log.Fatalf("Error while listening at port %s: %v\n", p.Port, err)
	}

	log.Println("Running reverse proxy on port", p.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error while accepting connection: ", err)
			continue
		}
		go p.handleConnection(conn)
	}
}
