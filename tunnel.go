package simtunnel

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	localhost = "localhost"
)

var (
	ErrCopyEmptyBuffer = fmt.Errorf("copy empty buffer")
	ErrClosedListener  = fmt.Errorf("closed listener")
)

func (tunnel *Tunnel) netCopy(input, output net.Conn) (err error) {
	var count int64
	count, err = io.Copy(output, input)
	if err == nil && count < 0 {
		err = ErrCopyEmptyBuffer
		return
	}
	return
}

// Tunnel struct
type Tunnel struct {
	srcAddr     string
	forwardAddr string
	connsSize   int
	srcListener net.Listener
	sleepTime   time.Duration
	forwardConns map[net.Conn]net.Conn
	od          sync.Once
	ch          chan struct{}
}

// NewTunnel returns tunnel
func NewTunnel(sleepTime time.Duration, connsSize int) (tunnel Tunnel) {
	tunnel.ch = make(chan struct{})
	tunnel.sleepTime = sleepTime
	tunnel.connsSize = connsSize
	tunnel.forwardConns = make(map[net.Conn]net.Conn, connsSize)
	return
}

func (tunnel *Tunnel) serveLn(ln net.Listener, forwardAddr string) (err error) {
	for {
		var conn, forwardConn net.Conn
		conn, err = ln.Accept()
		if err != nil {
			if tunnel.Closed() {
				err = ErrClosedListener
				return
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(tunnel.sleepTime)
			}
			continue
		}
		forwardConn, err = net.Dial("tcp", forwardAddr)
		if err != nil {
			conn.Close()
			continue
		}
		// keep connections
		tunnel.forwardConns[conn] = conn
		go func() {
			go tunnel.netCopy(conn, forwardConn)
			tunnel.netCopy(forwardConn, conn)
			forwardConn.Close()
			conn.Close()
			delete(tunnel.forwardConns, forwardConn)
		}()
	}
}

// ListenAndServe start local tunnel server
func (tunnel *Tunnel) ListenAndServe(srcPort int, forwardHost string, forwardPort int) (err error) {
	if tunnel.Closed() {
		return
	}
	if tunnel.srcListener != nil {
		return
	}
	tunnel.srcAddr = net.JoinHostPort(localhost, strconv.Itoa(srcPort))
	tunnel.forwardAddr = net.JoinHostPort(forwardHost, strconv.Itoa(forwardPort))
	tunnel.srcListener, err = net.Listen("tcp", tunnel.srcAddr)
	if err != nil {
		return
	}
	return tunnel.serveLn(tunnel.srcListener, tunnel.forwardAddr)
}

// Closed returns whether tunnel was closed
func (tunnel *Tunnel) Closed() bool {
	select {
	case <-tunnel.ch:
		return true
	default:
		return false
	}
}

// Close stop local tunnel server
func (tunnel *Tunnel) Close() {
	tunnel.od.Do(func() {
		close(tunnel.ch)
		if tunnel.srcListener != nil {
			tunnel.srcListener.Close()
		}
		// close all unclosed connections
		for _, c := range tunnel.forwardConns {
			c.Close()
		}
	})
}
