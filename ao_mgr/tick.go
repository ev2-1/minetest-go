package ao

import (
	"github.com/anon55555/mt"
)

var rmQueue []mt.AOID

func Tick() {
	// check if each client has all aos
	clientsMu.RLock()
	activeObjectsMu.RLock()

	for _, d := range clients {
		for id, _ := range activeObjects {
			if _, ok := d.aos[id]; !ok {
				// clt dosn't have AO, adding to queue:
				d.QueueAdd(id)
			}
		}
	}

	activeObjectsMu.RUnlock()
	clientsMu.RUnlock()
}
