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
	loadConfig()

	// initialize Logging
	initLog()

	// initialize ticks:
	initTicks()

	// open ClientDataDB
	if err := initClientDataDB(); err != nil {
		log.Fatalf("Error initializing Client Data DB: %s!", err)
	}

	var listenAddr = ":30000"
	if addr, ok := GetConfig("listen").(string); ok && addr != "" {
		listenAddr = addr
	}
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

		wg := sync.WaitGroup{}

		// Kicking all clients
		clts := Clts()

		log.Print("sending shutdown to all clients")
		for c := range clts {
			wg.Add(1)

			go func(c *Client) {
				ack, err := c.Kick(mt.Shutdown, "Shutting down.")
				if err == nil {
					<-ack
				}

				wg.Done()
			}(c)
		}

		shutdownHooksMu.RLock()
		defer shutdownHooksMu.RUnlock()

		log.Println("executing shutdown Hooks")

		for _, h := range shutdownHooks {
			go func(h func()) {
				wg.Add(1)
				h()
				wg.Done()
			}(h)
		}

		wg.Wait()

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
