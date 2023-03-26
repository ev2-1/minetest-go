package minetest

import (
	"github.com/anon55555/mt"
	"github.com/kevinburke/nacl/randombytes"

	"bytes"
	"fmt"
	"os"
	"time"
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
		for h := range leaveHooks {
			h.Thing(l)
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

// like log.Fatalf but does not panic but kick Client with specified Message
func (c *Client) Fatalf(str string, v ...any) (ack <-chan struct{}, err error) {
	return c.Kick(mt.Custom, fmt.Sprintf(str, v...))
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
		Loggers.Errorf("Failed to add new player to database: %s\n", 1, err)
		os.Exit(1)
	}

	c.UUID = id

	registerHooksMu.RLock()
	defer registerHooksMu.RUnlock()
	for h := range registerHooks {
		h.Thing(c)
	}

	return nil
}

// RegisterPlayer as active
func RegisterPlayer(c *Client) {
	clientsMu.Lock()
	clients[c] = struct{}{}
	clientsMu.Unlock()

	joinHooksMu.RLock()
	for h := range joinHooks {
		h.Thing(c)
	}
	joinHooksMu.RUnlock()

	close(c.initCh)
}

func InitClient(c *Client) {
	var err error

	// get UUID:
	c.UUID, err = DB_PlayerGetByName(c.Name)
	if err != nil || c.UUID == UUIDNil {
		c.Logf("Player joined for the first time! (err: %s)\n", err)

		err = firstJoin(c)
		if err != nil {
			c.Logf("Failed to Register: %s\n", err)

			c.SendCmd(&mt.ToCltKick{
				Reason: mt.SrvErr,
			})

			return
		}
	}

	c.Logf("Got UUID: %s\n", c.UUID)

	// data:
	var bytesTtl int
	c.data, bytesTtl, err = DB_PlayerGetData(c.UUID)
	if err != nil {
		c.Logf("Failed to get Player data!: %s\n", err)
		panic("Failed to get Player data!")
	}

	c.Logf("Loaded %d fields. a total of %d bytes\n", len(c.data), bytesTtl)

	//pos:
	clientPos, ok := c.data["pos"]
	if !ok {
		c.Pos = MakePos(c)
	} else {
		dat := clientPos.(*ClientDataSaved)
		c.Pos = new(ClientPos)

		err := c.Pos.Deserialize(bytes.NewReader(dat.Bytes()))
		if err != nil {
			c.Logf("Error while Deserializing ClientPos: %s\n", err)
			c.Pos = MakePos(c)
		}
	}

	c.AOData = makeAOData()

	c.openSpecsT = make(map[string]time.Time)
	c.openSpecs = make(map[string]*Formspec)
	loaderName := DefaultMapLoader()
	if loaderName == "" {
		c.SendCmd(&mt.ToCltKick{
			Reason: mt.SrvErr,
		})

		c.Logf("DefaultMapLoader is empty")
		return
	}

	mapLoader := GetMapLoader(loaderName)
	if mapLoader == nil {
		c.SendCmd(&mt.ToCltKick{
			Reason: mt.SrvErr,
		})

		c.Logf("DefaultMapLoader is not registerd \"%s\"", loaderName)
		return
	}

	c.SetMapLoader(mapLoader)

	initHooksMu.RLock()
	defer initHooksMu.RUnlock()

	for h := range initHooks {
		h.Thing(c)
	}
}
