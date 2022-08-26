package minetest

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/anon55555/mt"
)

var runOnce sync.Once

// Starts the server
func Run() {
	runOnce.Do(runFunc)
}

func runFunc() {
	parseArguments()

	// initialize Logging
	initLog()

	// initialize ticks:
	initTicks()

	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	pc, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	l := listen(pc)
	defer l.Close()

	log.Println("listen", l.Addr())

	// killchan handeling
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		<-sig

		log.Print("received SIGINT or other Interrupt")

		go func() {
			clts := Clts()

			log.Print("sending shutdown to all clients")
			for c := range clts {
				go c.Kick(mt.Shutdown, "Shutting down.")
			}
		}()

		time.Sleep(time.Second * 1)

		os.Exit(0)
	}()

	for {
		c, err := l.accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Print("stop listening")
				continue
			}

			log.Print(err)
			continue
		}

		go func() {
			<-c.Init()
			c.Log("<->", "handshake complete")
		}()
	}
}
