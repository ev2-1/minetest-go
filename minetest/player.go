package minetest

import (
	"github.com/anon55555/mt"

	"fmt"
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

func CltLeave(l *Leave) {
	l.Client.leaveOnce.Do(func() {
		for _, h := range leaveHooks {
			h(l)
		}
	})

	clientsMu.Lock()
	delete(clients, l.Client)
	clientsMu.Unlock()

	clientsMu.RLock()
	defer clientsMu.RUnlock()

	// Do not send clt kick if disconnected by self
	if l.AbstrReason == Kick {
		cmd := &mt.ToCltKick{
			Reason: l.Reason,
			Custom: l.Custom,
		}

		l.Client.SendCmd(cmd)
	}
}

func Clts() map[*Client]struct{} {
	clientsMu.RLock()
	defer clientsMu.RLock()

	c := make(map[*Client]struct{}, len(clients))

	for k, v := range clients {
		c[k] = v
	}

	return c
}

func (c *Client) Kick(r mt.KickReason, Custom string) {
	CltLeave(&Leave{
		Client: c,
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

func registerPlayer(c *Client) {
	for _, h := range joinHooks {
		h(c)
	}

	// change prefix to new name
	c.Logger.SetPrefix(fmt.Sprintf("[%s %s] ", c.RemoteAddr(), c.Name))

	clientsMu.Lock()
	defer clientsMu.Unlock()

	clients[c] = struct{}{}
}

func InitClient(c *Client) {
	for _, h := range initHooks {
		h(c)
	}
}
