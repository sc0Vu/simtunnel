package simtunnel

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
)

const (
	localhost = "localhost"
)

var (
	ErrCopyEmptyBuffer = fmt.Errorf("copy empty buffer")
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
	forwardConn net.Conn
	srcAddr     string
	forwardAddr string
	srcListener net.Listener
	started     bool
	mx          sync.Mutex
}

// NewTunnel returns tunnel
func NewTunnel(srcPort int, forwardHost string, forwardPort int) (tunnel Tunnel) {
	tunnel.srcAddr = net.JoinHostPort(localhost, strconv.Itoa(srcPort))
	tunnel.forwardAddr = net.JoinHostPort(forwardHost, strconv.Itoa(forwardPort))
	return
}

// ListenAndServe start local tunnel server
func (tunnel *Tunnel) ListenAndServe() (err error) {
	tunnel.mx.Lock()
	defer tunnel.mx.Unlock()
	if tunnel.started {
		return
	}
	tunnel.srcListener, err = net.Listen("tcp", tunnel.srcAddr)
	if err != nil {
		return
	}
	tunnel.forwardConn, err = net.Dial("tcp", tunnel.forwardAddr)
	if err != nil {
		return
	}
	tunnel.started = true
	go func() {
		for {
			var conn net.Conn
			conn, err = tunnel.srcListener.Accept()
			if err != nil {
				continue
			}
			go tunnel.netCopy(tunnel.forwardConn, conn)
			tunnel.netCopy(conn, tunnel.forwardConn)
			conn.Close()
		}
	}()
	return
}

// Close stop local tunnel server
func (tunnel *Tunnel) Close() (err error) {
	tunnel.mx.Lock()
	defer tunnel.mx.Unlock()
	if !tunnel.started {
		return
	}
	err = tunnel.srcListener.Close()
	if err != nil {
		return
	}
	err = tunnel.forwardConn.Close()
	if err != nil {
		return
	}
	tunnel.started = false
	return
}
