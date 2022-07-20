package minetest

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/anon55555/mt"
	"github.com/anon55555/mt/rudp"
)

type ClientState uint8

const (
	CsCreated ClientState = iota
	CsInit
	CsActive
	CsSudo
)

// A Client represents a client
type Client struct {
	mt.Peer
	mu sync.RWMutex

	Logger *log.Logger

	State   ClientState
	Name    string
	initCh  chan struct{}
	aoReady sync.Once

	leaveOnce sync.Once // a client only can leave once

	lang string
}

func (c *Client) SendCmd(cmd mt.Cmd) (ack <-chan struct{}, err error) {
	if verbose {
		c.Log("<-", fmt.Sprintf("%T", cmd))
	}

	return c.Peer.SendCmd(cmd)
}

func (c *Client) SetState(state ClientState) {
	//	oldState := c.State
	c.State = state

	//	if oldState != state {
	//		updateState(c, oldState, state)
	//	}
}

func (c *Client) Log(dir string, v ...any) {
	c.Logger.Println(append([]any{dir}, v...)...)
}

func handleClt(c *Client) {
	for {
		pkt, err := c.Recv()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				if errors.Is(c.WhyClosed(), rudp.ErrTimedOut) {
					c.Log("<->", "timeout")
					CltLeave(&Leave{
						Client:      c,
						AbstrReason: Timeout,
					})
				} else {
					CltLeave(&Leave{
						Client:      c,
						AbstrReason: Exit,
					})
					c.Log("<->", "disconnect")
				}

				if c.Name != "" {
					CltLeave(&Leave{
						Client:      c,
						AbstrReason: NetErr,
					})
				}

				break
			}

			c.Log("->", err)
			continue
		}

		c.process(&pkt)
	}
}

func (c *Client) Init() <-chan struct{} { return c.initCh }
