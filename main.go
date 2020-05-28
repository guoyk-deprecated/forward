package main

import (
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	optBind   = os.Getenv("BIND")
	optTarget = os.Getenv("TARGET")
)

func exit(err *error) {
	if *err != nil {
		log.Printf("exited with error: %s", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	var err error
	defer exit(&err)

	if len(optBind) == 0 {
		err = errors.New("$BIND not specified")
		return
	}

	if len(optTarget) == 0 {
		err = errors.New("$TARGET not specified")
		return
	}

	if _, err = net.ResolveTCPAddr("tcp", optBind); err != nil {
		return
	}

	if _, err = net.ResolveTCPAddr("tcp", optTarget); err != nil {
		return
	}

	log.Printf("forward started, %s -> %s", optBind, optTarget)

	var l net.Listener
	if l, err = net.Listen("tcp", optBind); err != nil {
		return
	}
	defer l.Close()

	done := make(chan interface{})
	go func() {
		for {
			var err error
			var c net.Conn
			if c, err = l.Accept(); err != nil {
				log.Printf("failed to accept connection: %s", err.Error())
				break
			}
			go handle(c)
		}
		close(done)
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-done:
		err = errors.New("listener closed unexpectedly")
	case sig := <-sigCh:
		log.Printf("signal received: %s", sig.String())
	}
}

func handle(c net.Conn) {
	var err error

	defer c.Close()

	var t net.Conn
	if t, err = net.Dial("tcp", optTarget); err != nil {
		log.Printf("failed to dial %s for %s: %s", optTarget, c.RemoteAddr().String(), err.Error())
		return
	}
	defer t.Close()

	go io.Copy(c, t)
	io.Copy(t, c)
}
