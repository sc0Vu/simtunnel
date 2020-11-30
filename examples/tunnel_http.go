package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sc0vu/simtunnel"
)

func main() {
	localPort := 9999
	forwardPort := 8080
	tunnel := simtunnel.NewTunnel(10*time.Millisecond, 1024)
	go func() {
		err := tunnel.ListenAndServe(localPort, "localhost", forwardPort)
		if err != nil && err != simtunnel.ErrClosedListener {
			log.Printf("Failed to listen and serve the tunnel: %s\n", err.Error())
		}
	}()
	log.Printf("Listen and serve on port: %d\n", localPort)
	log.Printf("Forward to: localhost:%d\n", forwardPort)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	sig := <-sigChan
	switch sig {
	case syscall.SIGINT:
		break
	}
	tunnel.Close()
	log.Printf("Close tunnel")
}
