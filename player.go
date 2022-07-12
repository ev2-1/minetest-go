package minetest

import (
	"github.com/anon55555/mt"
)

type Leave struct {
	Reason      mt.KickReason
	Custom      string
	AbstrReason Reason

	Client *Client
}

func (l *Leave) pkt() *mt.ToCltKick {
	return &mt.ToCltKick{
		Reason: l.Reason,
		Custom: l.Custom,
	}
}

type Reason uint8

const (
	Kick Reason = iota
	Timeout
	Exit
	NetErr
)

var leaveChan = make(chan *Leave)

func LeaveChan() <-chan *Leave {
	return leaveChan
}

func CltLeave(l *Leave) {
	clientsMu.Lock()
	delete(clients, l.Client)
	clientsMu.Unlock()

	clientsMu.RLock()
	defer clientsMu.RUnlock()

	ack, _ := l.Client.SendCmd(&mt.ToCltKick{
		Reason: l.Reason,
		Custom: l.Custom,
	})

	leaveChan <- l

	select {
	case <-l.Client.Closed():
	case <-ack:
		l.Client.Close()
	}
}

func Clts() map[*Client]struct{} {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	return clients
}

func (c *Client) Kick(r mt.KickReason, Custom string) {
	CltLeave(&Leave{
		Reason: r,
		Custom: Custom,

		AbstrReason: Kick,
	})
}

func PlayerByName(name string) *Client {
	for c := range Clts() {
		if c.Name == name {
			return c
		}
	}

	return nil
}

func PlayerExists(name string) bool {
	return PlayerByName(name) != nil
}

func RegisterPlayer(c *Client) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	joinCh <- c
	clients[c] = struct{}{}
}

var joinCh = make(chan *Client)

func JoinChan() <-chan *Client {
	return joinCh
}
