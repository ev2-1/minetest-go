package minetest

import (
	"errors"
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
		Loggers.Errorf("Error initializing Client Data DB: %s!", 1, err)
		os.Exit(1)
	}

	listenAddr, _ := GetConfig("listen", ":30000")
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		Loggers.Errorf("udp resolve: %s", 1, err)
		os.Exit(1)
	}

	pc, err := net.ListenUDP("udp", addr)
	if err != nil {
		Loggers.Errorf("%s", 1, err)
		os.Exit(1)
	}

	l := listen(pc)
	defer l.Close()

	Loggers.Defaultf("listening on %s\n", 1, l.Addr())

	// set state to Online
	stateMu.Lock()
	state = StateOnline
	stateMu.Unlock()

	// killchan handeling
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		<-sig

		Loggers.Defaultf("received SIGINT or other Interrupt - Shutting Down\n", 1)
		stateMu.Lock()
		state = StateShuttingDown
		stateMu.Unlock()

		wg := sync.WaitGroup{}

		// Kicking all clients
		clts := Clts()

		Loggers.Defaultf("sending shutdown to all clients\n", 1)
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

		Loggers.Defaultf("executing shutdown Hooks\n", 1)

		for h := range shutdownHooks {
			wg.Add(1)
			go func(h ShutdownHook) {
				h()
				wg.Done()
			}(h.Thing)
		}

		wg.Wait()
		wg = sync.WaitGroup{}

		Loggers.Defaultf("executing savefile Hooks\n", 1)
		for h := range saveFileHooks {
			wg.Add(1)
			go func(h SaveFileHook) {
				h()
				wg.Done()
			}(h.Thing)
		}

		wg.Wait()

		Loggers.Defaultf("os.Exit(0)\n", 1)
		os.Exit(0)
	}()

	for {
		c, err := l.accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				Loggers.Defaultf("stop listening\n", 1)
				continue
			}

			Loggers.Defaultf("%s", 1, err)
			continue
		}

		go func() {
			<-c.Init()
			c.Log("<->", "handshake complete")
		}()
	}
}
