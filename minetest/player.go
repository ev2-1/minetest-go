package minetest

import (
	"github.com/anon55555/mt"
	"github.com/kevinburke/nacl/randombytes"

	"fmt"
	"log"
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

func CltLeave(l *Leave) (ack <-chan struct{}, err error) {
	l.Client.leaveOnce.Do(func() {
		leaveHooksMu.RLock()
		for _, h := range leaveHooks {
			h(l)
		}

		leaveHooksMu.RUnlock()

		SyncPlayerData(l.Client)
	})

	clientsMu.Lock()
	delete(clients, l.Client)
	clientsMu.Unlock()

	// Do not send clt kick if disconnected by self
	if l.AbstrReason == Kick {
		cmd := &mt.ToCltKick{
			Reason: l.Reason,
			Custom: l.Custom,
		}

		return l.Client.SendCmd(cmd)
	}

	aack := make(chan struct{})
	close(aack)

	return aack, nil
}

func Clts() map[*Client]struct{} {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	c := make(map[*Client]struct{}, len(clients))

	for k, v := range clients {
		c[k] = v
	}

	return c
}

func (c *Client) Kick(r mt.KickReason, Custom string) (ack <-chan struct{}, err error) {
	return CltLeave(&Leave{
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

func genUUID() (u UUID) {
	uid := make([]byte, 16)

	for i := 0; i < 100; i++ {
		randombytes.MustRead(uid)
		copy(u[:], uid)

		if u == UUIDNil {
			continue
		}

		name, err := DB_PlayerGetByUUID(u)
		if err == nil && name == "" {
			return
		}
	}

	panic("Cant generate UUID!, 100 tries! all registerd!")
}

func firstJoin(c *Client) error {
	id := genUUID()

	c.Logf("Generated UUID %s for Player %s", id, c.Name)
	if err := DB_PlayerSet(id, c.Name); err != nil {
		log.Fatalf("Failed to add new player to database: %s\n", err)

		return err
	}

	c.UUID = id

	registerHooksMu.RLock()
	defer registerHooksMu.RUnlock()
	for _, h := range registerHooks {
		h(c)
	}

	return nil
}

// register Player as active
func registerPlayer(c *Client) {
	clientsMu.Lock()
	clients[c] = struct{}{}
	clientsMu.Unlock()

	joinHooksMu.RLock()
	for _, h := range joinHooks {
		h(c)
	}
	joinHooksMu.RUnlock()

	// change prefix to new name
	c.Logger.SetPrefix(fmt.Sprintf("[%s %s] ", c.RemoteAddr(), c.Name))
}

func InitClient(c *Client) {
	var err error

	// get UUID:
	c.UUID, err = DB_PlayerGetByName(c.Name)
	if err != nil || c.UUID == UUIDNil {
		c.Logf("Player joined for the first time! (err: %s)\n", err)

		err = firstJoin(c)
		if err != nil {
			c.Log("Failed to Register: %s\n", err)

			c.SendCmd(&mt.ToCltKick{
				Reason: mt.SrvErr,
			})

			return
		}
	}

	c.Logf("Got UUID: %s\n", c.UUID)

	// data:
	var bytes int
	c.data, bytes, err = DB_PlayerGetData(c.UUID)
	if err != nil {
		c.Log("Failed to get Player data!: %s\n", err)
		panic("Failed to get Player data!")
	}

	c.Logf("Loaded %d fields. a total of %d bytes\n", len(c.data), bytes)

	initHooksMu.RLock()
	defer initHooksMu.RUnlock()

	for _, h := range initHooks {
		h(c)
	}
}
