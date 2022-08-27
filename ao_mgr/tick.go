package ao

import (
	"github.com/ev2-1/minetest-go/minetest"

	"github.com/anon55555/mt"
)

var rmQueue []mt.AOID

func init() {
	minetest.RegisterTickHook(func() {
		// check if each client has all aos
		clientsMu.RLock()
		activeObjectsMu.RLock()

		for _, d := range clients {
			d.aosMu.RLock()

			for id, _ := range activeObjects {
				if _, ok := d.aos[id]; !ok {
					// clt dosn't have AO, adding to queue:
					d.QueueAdd(id)
				}
			}

			d.aosMu.RUnlock()
		}

		activeObjectsMu.RUnlock()
		clientsMu.RUnlock()
	})
}
