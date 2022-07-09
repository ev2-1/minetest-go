package minetest

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/anon55555/mt"
)

var runOnce sync.Once

// Starts the server
func Run() {
	runOnce.Do(runFunc)
}

func runFunc() {
	parseArguments()

	// load plugins
	loadPlugins()
	initLog()

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

		clts := Clts()

		var wg sync.WaitGroup
		wg.Add(len(clts))

		for c := range clts {
			go func(c *Client) {
				c.Kick(mt.Shutdown, "Shutting down.")
				<-c.Closed()
				wg.Done()
			}(c)
		}

		wg.Wait()
		os.Exit(0)
	}()

	for {
		c, err := l.accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Print("stop listening")
				break
			}

			log.Print(err)
			break
		}

		go func() {
			<-c.Init()
			c.Log("<->", "handshake complete")

			// TODO: actually do it:
			//connect(conn, c)
		}()
	}
}
