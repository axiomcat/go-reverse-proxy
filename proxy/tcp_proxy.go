package proxy

import (
	"fmt"
	"io"
	"net"

	"github.com/axiomcat/reverse-proxy/logger"
)

type TcpProxy struct {
	Port       string
	TargetAddr string
}

func (p TcpProxy) handleConnection(conn net.Conn) {
	logger := logger.GetInstance(0)
	connAddress := fmt.Sprintf("%v:%v", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	logger.Debug(fmt.Sprintf("Handling connection %s\n", connAddress))

	targetConn, err := net.Dial("tcp", p.TargetAddr)
	if err != nil {
		conn.Close()
		logger.Fatal(fmt.Sprintf("Can't connect to %v: %v", p.TargetAddr, err))
	}

	logger.Debug("Copying connecton to target")
	go io.Copy(targetConn, conn)
	logger.Debug("Copying connecton from target")
	go io.Copy(conn, targetConn)
}

func (p TcpProxy) Start() {
	logger := logger.GetInstance(0)

	ln, err := net.Listen("tcp", p.Port)
	defer ln.Close()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Error while listening at port %s: %v\n", p.Port, err))
	}

	logger.Log(fmt.Sprint("Running TCP reverse proxy on port", p.Port))

	for {
		conn, err := ln.Accept()

		logger.Log(fmt.Sprint("Got TCP connection", conn))

		if err != nil {
			logger.Log(fmt.Sprint("Error while accepting connection: ", err))
			continue
		}

		go p.handleConnection(conn)
	}
}
