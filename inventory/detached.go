package inventory

import (
	"github.com/ev2-1/minetest-go/minetest"

	//sync "github.com/sasha-s/go-deadlock"
	"sync"
)

var (
	detachedInvs   = make(map[string]*DetachedInv)
	detachedInvsMu sync.RWMutex
)

type DetachedInv struct {
	SimpleInv
}

func GetDetached(name string, c *minetest.Client) (inv *DetachedInv, err error) {
	detachedInvsMu.RLock()
	defer detachedInvsMu.RUnlock()

	c.Log("[INV] access detached inv '%s'", name)

	inv, ok := detachedInvs[name]
	if !ok {
		return nil, ErrInvalidInv
	}

	return
}

func RegisterDetached(name string, inv *DetachedInv) {
	detachedInvsMu.Lock()
	defer detachedInvsMu.Unlock()

	detachedInvs[name] = inv
}

var _ RWInv = &DetachedInv{}
