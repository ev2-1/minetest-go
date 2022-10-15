package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"bytes"
	"sync"
)

var (
	playerInventories   = make(map[*minetest.Client]*PlayerInv)
	playerInventoriesMu sync.RWMutex
)

// GetPlayerInv returns a pointer to the players inventory
// can return <nil>!
func GetPlayerInv(c *minetest.Client) *PlayerInv {
	playerInventoriesMu.RLock()
	defer playerInventoriesMu.RUnlock()

	return playerInventories[c]
}

func init() {
	minetest.RegisterJoinHook(func(c *minetest.Client) {
		playerInventoriesMu.Lock()
		defer playerInventoriesMu.Unlock()

		maker, ok := playerInventoryTypes[defaultInventory]
		if !ok {
			return
		}

		inv := maker(c)

		// TODO read invtype from somewhere
		ack, _ := c.SendCmd(&mt.ToCltInvFormspec{
			Formspec: inv.Formspec,
		})

		buf := &bytes.Buffer{}
		inv.Serialize(buf)

		<-ack

		c.SendCmd(&mt.ToCltInv{
			Inv: buf.String(),
		})

		playerInventories[c] = &inv
	})
}
