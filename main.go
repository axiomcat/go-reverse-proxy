package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	connAddress := fmt.Sprintf("%v:%v", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	log.Printf("Received connection %s\n", connAddress)

	targetAddr := "localhost:8080"
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		conn.Close()
		log.Fatalf("Can't connect to %v: %v", targetAddr, err)
	}

	go io.Copy(targetConn, conn)
	go io.Copy(conn, targetConn)
}

func main() {
	addr := ":8020"
	ln, err := net.Listen("tcp", addr)
	defer ln.Close()

	if err != nil {
		log.Fatalf("Error while listening at port %s: %v\n", addr, err)
	}

	log.Println("Running reverse proxy on port", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error while accepting connection: ", err)
			continue
		}
		go handleConnection(conn)
	}
}
