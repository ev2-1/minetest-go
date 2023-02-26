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

	// create ItemDef net cache
	makeItemNetCache()

	// initialize ticks:
	initTicks()

	// open ClientDataDB
	if err := initClientDataDB(); err != nil {
		log.Fatalf("Error initializing Client Data DB: %s!", err)
	}

	listenAddr, _ := GetConfig("listen", ":30000")
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

	// set state to Online
	stateMu.Lock()
	state = StateOnline
	stateMu.Unlock()

	// killchan handeling
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		<-sig

		log.Print("received SIGINT or other Interrupt - Shutting Down")
		stateMu.Lock()
		state = StateShuttingDown
		stateMu.Unlock()

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

		for h := range shutdownHooks {
			wg.Add(1)
			go func(h ShutdownHook) {
				h()
				wg.Done()
			}(h.Thing)
		}

		wg.Wait()
		wg = sync.WaitGroup{}

		log.Println("executing savefile Hooks")
		for h := range saveFileHooks {
			wg.Add(1)
			go func(h SaveFileHook) {
				h()
				wg.Done()
			}(h.Thing)
		}

		wg.Wait()

		log.Printf("os.Exit(0)")
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
