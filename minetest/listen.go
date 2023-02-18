package minetest

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/anon55555/mt"
)

var (
	ErrServerShuttingDown = errors.New("Server shutting down")
)

type listener struct {
	mt.Listener

	mu sync.RWMutex

	clts []*Client
}

var listeners []*listener
var listenersMu sync.RWMutex

var clients = make(map[*Client]struct{})
var clientsMu sync.RWMutex

func listen(pc net.PacketConn) *listener {
	l := &listener{
		Listener: mt.Listen(pc),
		// clts is array and dosnt need to be initialized
	}

	listenersMu.Lock()
	defer listenersMu.Unlock()

	listeners = append(listeners, l)

	return l
}

func (l *listener) Clts() []*Client {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.clts
}

func (l *listener) accept() (*Client, error) {
	p, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// if not online reject
	if State() != StateOnline {
		p.SendCmd(&mt.ToCltKick{
			Reason:    mt.Shutdown,
			Reconnect: true,
		})

		log.Printf("WARN: %s tried to connect but state 1= StateOnline!\n", p.RemoteAddr())

		return nil, ErrServerShuttingDown
	}

	c := &Client{
		Peer:   p,
		initCh: make(chan struct{}),
	}

	c.Logger = log.New(logWriter, c.String(), logFlags)

	l.mu.Lock()
	l.clts = append(l.clts, c)
	l.mu.Unlock()

	c.Log("->", "connect")
	go handleClt(c)

	select {
	case <-c.Closed():
		return nil, fmt.Errorf("%s is closed", c.RemoteAddr())
	default:
	}

	return c, nil
}
