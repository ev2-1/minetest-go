package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	//sync "github.com/sasha-s/go-deadlock"
	"log"
	"sync"
)

var (
	detachedInvs   = make(map[string]*DetachedInv)
	detachedInvsMu sync.RWMutex
)

type DetachedInv struct {
	SimpleInv
	Name string

	// List of clients suppost to have access
	ClientsMu sync.RWMutex
	Clients   map[*minetest.Client]struct{}
}

func GetDetached(name string, c *minetest.Client) (inv *DetachedInv, err error) {
	detachedInvsMu.RLock()
	defer detachedInvsMu.RUnlock()

	c.Logger.Printf("[INV] access detached inv '%s'", name)

	inv, ok := detachedInvs[name]
	if !ok {
		return nil, ErrInvalidInv
	}

	return
}

func (di *DetachedInv) Set(k string, v InvList) {
	di.SimpleInv.Set(k, v)

	// Update:
	_, err := di.SendUpdates()
	if err != nil {
		log.Printf("Error occured while updating DetachedInv: %s\n", err)
	}
}

func (di *DetachedInv) SendUpdates() (<-chan struct{}, error) {
	di.ClientsMu.RLock()
	defer di.ClientsMu.RUnlock()

	str, err := SerializeString(di.Serialize)
	if err != nil {
		return nil, err
	}

	var acks []<-chan struct{}

	for c := range di.Clients {
		ack, err := c.SendCmd(&mt.ToCltDetachedInv{
			Name: di.Name,
			Keep: true,

			Inv: str,
		})

		if err != nil {
			continue
		}

		acks = append(acks, ack)
	}

	return minetest.Acks(acks...), nil
}

func (di *DetachedInv) AddClient(c *minetest.Client) (<-chan struct{}, error) {
	di.ClientsMu.Lock()
	defer di.ClientsMu.Unlock()

	str, err := SerializeString(di.Serialize)
	if err != nil {
		c.Logger.Printf("Error: %s", err)
		return nil, err
	}

	// send detached inv to test:
	ack, err := c.SendCmd(&mt.ToCltDetachedInv{
		Name: di.Name,
		Keep: true,

		Inv: str,
	})

	di.Clients[c] = struct{}{}

	return ack, err
}

func (di *DetachedInv) RmClient(c *minetest.Client) {
	di.ClientsMu.Lock()
	defer di.ClientsMu.Unlock()

	delete(di.Clients, c)
}

func RegisterDetached(name string, inv *DetachedInv) {
	detachedInvsMu.Lock()
	defer detachedInvsMu.Unlock()
	inv.Name = name

	if inv.Clients == nil {
		inv.Clients = make(map[*minetest.Client]struct{})
	}

	detachedInvs[name] = inv
}

var _ RWInv = &DetachedInv{}
