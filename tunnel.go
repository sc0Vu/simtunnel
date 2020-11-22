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
	srcListener net.Listener
	od          sync.Once
	ch          chan struct{}
}

// NewTunnel returns tunnel
func NewTunnel() (tunnel Tunnel) {
	tunnel.ch = make(chan struct{})
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
				sleepTime := 10 * time.Millisecond
				time.Sleep(sleepTime)
			}
			continue
		}
		forwardConn, err = net.Dial("tcp", forwardAddr)
		if err != nil {
			return
		}
		go func() {
			go tunnel.netCopy(conn, forwardConn)
			tunnel.netCopy(forwardConn, conn)
			forwardConn.Close()
			conn.Close()
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
	})
}
