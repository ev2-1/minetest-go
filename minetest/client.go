package minetest

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

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

type UUID [16]byte

// Returns UUID.RFC4122
func (u UUID) String() string {
	return u.RFC4122()
}

// Returns a RFC4122 formatted UUID
func (u UUID) RFC4122() string {
	//4 2 2 2 6
	return fmt.Sprintf("%x-%x-%x-%x", u[0:3], u[4:5], u[6:7], u[8:16])
}

// Returns UUID in 4 hex-encoded 4byte blocks seperated by dashes
func (u UUID) Quaters() string {
	return fmt.Sprintf("%x-%x-%x-%x", u[0:3], u[4:8], u[9:12], u[13:16])
}

var UUIDNil UUID

// A Client represents a client
type Client struct {
	mt.Peer
	mu sync.RWMutex

	UUID UUID
	Name string

	State   ClientState
	initCh  chan struct{}
	aoReady sync.Once

	data   map[string]ClientData
	dataMu sync.RWMutex

	leaveOnce sync.Once // a client only can leave once

	Pos      *ClientPos
	PosState sync.RWMutex

	lang string

	diggingMu    sync.RWMutex
	digging      *IntPos
	startDigging time.Time

	//formspecs
	openSpecsMu sync.RWMutex
	openSpecs   map[string]time.Time

	//         set by     registerd
	mapLoader *Registerd[*Registerd[MapLoader]]
}

func (c *Client) MapLoad() bool {
	loader := c.GetMapLoader()

	if loader == nil || loader.Thing == nil {
		return false
	}

	loader.Thing.Thing.Load()

	return true
}

func (c *Client) GetMapLoader() *Registerd[*Registerd[MapLoader]] {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.mapLoader
}

func (c *Client) SetMapLoader(loader *Registerd[MapLoader]) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.mapLoader = &Registerd[*Registerd[MapLoader]]{
		Caller(1),
		&Registerd[MapLoader]{
			loader.Path(),
			loader.Thing.Make(c),
		},
	}
}

func (c *Client) IsDigging() bool {
	c.diggingMu.RLock()
	defer c.diggingMu.RUnlock()

	return c.digging != nil
}

// Returns nil if *Client is not digging
func (c *Client) DigPos() (*IntPos, time.Time) {
	c.diggingMu.RLock()
	defer c.diggingMu.RUnlock()

	return c.digging, c.startDigging
}

// Returns nil if *Client is not digging
func (c *Client) setDigPos(p *IntPos) {
	c.diggingMu.Lock()
	defer c.diggingMu.Unlock()

	c.digging = p
	c.startDigging = time.Now()
}

func (c *Client) String() string {
	return fmt.Sprintf("[%s%s] ", c.Peer.RemoteAddr(), T(c.Name == "", "", " "+c.Name))
}

func (c *Client) Logf(format string, v ...any) {
	Loggers.Defaultf(c.String()+format, 2, v...)
}

func (c *Client) SendCmd(cmd mt.Cmd) (ack <-chan struct{}, err error) {
	if ConfigVerbose() {
		switch cmd.(type) {
		case *mt.ToCltBlkData:
			break

		default:
			lpkts, ok := GetConfig("log-packets", false)

			if (ConfigVerbose() && !(ok && !lpkts)) || lpkts {
				c.Log("<-", fmt.Sprintf("%T", cmd))
			}
		}
	}

	// packet preprocessor
	packetPreMu.RLock()
	defer packetPreMu.RUnlock()

	for pre := range packetPre {
		if pre.Thing(c, cmd) {
			return nil, nil
		}
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

func (c *Client) Log(v ...any) {
	Loggers.Default(c.String()+fmt.Sprint(v...), 2)
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

// BroadcastClientS broadcasts a mt.Cmd to a client slice
func BroadcastClientS(s []*Client, cmd mt.Cmd) <-chan struct{} {
	var acks []<-chan struct{}

	for _, clt := range s {
		ack, _ := clt.SendCmd(cmd)
		acks = append(acks, ack)
	}

	return Acks(acks...)
}

// BroadcastClientM broadcasts a mt.Cmd to a client slice
func BroadcastClientM(s map[*Client]struct{}, cmd mt.Cmd) <-chan struct{} {
	var acks []<-chan struct{}

	for clt := range s {
		ack, _ := clt.SendCmd(cmd)
		acks = append(acks, ack)
	}

	return Acks(acks...)
}

// Combine acks into one ack
// Waits for all acks to close then closes ack
func Acks(acks ...<-chan struct{}) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		for _, ack := range acks {
			if ack == nil {
				continue
			}

			<-ack
		}

		close(ch)
	}()

	return ch
}

func T[K any](c bool, t, f K) K {
	if c {
		return t
	} else {
		return f
	}
}
