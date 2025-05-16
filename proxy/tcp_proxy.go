package proxy

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/axiomcat/reverse-proxy/logger"
	"github.com/axiomcat/reverse-proxy/metrics"
)

type connPair struct {
	client net.Conn
	target net.Conn
}

type TcpProxy struct {
	Port        string
	TargetAddr  string
	listener    net.Listener
	wg          sync.WaitGroup
	connMutex   sync.Mutex
	connections map[net.Conn]connPair
}

func (p *TcpProxy) handleConnection(conn net.Conn) {
	logger := logger.GetInstance(0)
	metrics := metrics.GetInstance()

	targetConn, err := net.Dial("tcp", p.TargetAddr)

	metrics.ActiveTcpConnections += 1
	metrics.TotalTcpConnections += 1

	if err != nil {
		conn.Close()
		logger.Fatal(fmt.Sprintf("Can't connect to %v: %v", p.TargetAddr, err))
	}

	p.connMutex.Lock()
	p.connections[conn] = connPair{client: conn, target: targetConn}
	p.connMutex.Unlock()

	defer func() {
		logger.Debug("Connection finished, cleaning up")

		p.connMutex.Lock()
		delete(p.connections, conn)
		p.connMutex.Unlock()

		conn.Close()
		targetConn.Close()
		metrics.ActiveTcpConnections -= 1
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(targetConn, conn)
		targetConn.Close()
	}()

	go func() {
		defer wg.Done()
		io.Copy(conn, targetConn)
		conn.Close()
	}()

	wg.Wait()
}

func (p *TcpProxy) Start() {
	logger := logger.GetInstance(0)
	p.connections = make(map[net.Conn]connPair)

	ln, err := net.Listen("tcp", p.Port)
	defer ln.Close()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Error while listening at port %s: %v\n", p.Port, err))
	}

	p.listener = ln

	logger.Log(fmt.Sprint("Running TCP reverse proxy on port", p.Port))

	for {
		conn, err := ln.Accept()

		logger.Log(fmt.Sprint("Got TCP connection", conn))

		if err != nil {
			logger.Log(fmt.Sprint("Listener closed: ", err))
			break
		}

		p.wg.Add(1)
		go func(c net.Conn) {
			defer p.wg.Done()
			p.handleConnection(c)
		}(conn)
	}
}

func (p *TcpProxy) Stop() {
	logger := logger.GetInstance(0)
	logger.Log("Shutting down Tcp proxy")

	if p.listener != nil {
		logger.Log("Closing tcp listener")
		p.listener.Close()
	}

	p.listener = nil

	p.connMutex.Lock()
	connPairs := make([]connPair, 0, len(p.connections))
	for _, pair := range p.connections {
		connPairs = append(connPairs, pair)
	}
	p.connMutex.Unlock()

	logger.Debug(fmt.Sprintf("Closing %d active connections", len(connPairs)))
	for _, conn := range connPairs {
		logger.Debug(fmt.Sprintf("Closing client: %v", conn.client.RemoteAddr()))
		conn.client.Close()
		logger.Debug(fmt.Sprintf("Closing target: %v", conn.target.RemoteAddr()))
		conn.target.Close()
	}

	p.wg.Wait()

	logger.Log("Tcp proxy shutdown gracefully")
}
