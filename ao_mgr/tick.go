package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"sync"
)

var rmQueue = make(map[mt.AOID]struct{})
var rmQueueMu sync.RWMutex

func init() {
	minetest.RegisterTickHook(func() {
		// check if each client has all aos
		clientsMu.RLock()
		activeObjectsMu.RLock()

		for _, d := range clients {
			d.aosMu.Lock()

			for id, _ := range activeObjects {
				if _, ok := d.aos[id]; !ok {
					// clt dosn't have AO, adding to queue:
					d.QueueAdd(id)
				}
			}

			d.aosMu.Unlock()
		}

		activeObjectsMu.RUnlock()
		rmQueueMu.Lock()

		// remove global remove queue
		if len(rmQueue) != 0 {
			for _, d := range clients {
				d.aosMu.RLock()

				for id, _ := range d.aos {
					if _, ok := rmQueue[id]; ok {
						// clt has stuff from rmqueue:
						d.QueueRm(id)
					}
				}

				d.aosMu.RUnlock()
			}

		}
		clientsMu.RUnlock()

		// apply rm to global aos
		if len(rmQueue) != 0 {
			activeObjectsMu.Lock()
			for id := range rmQueue {
				if _, ok := activeObjects[id]; ok {
					log.Printf("removing AO %d globally\n", id)
					delete(activeObjects, id)
				}
			}

			activeObjectsMu.Unlock()
		}

		// clear queue:
		rmQueue = map[mt.AOID]struct{}{}

		rmQueueMu.Unlock()
	})
}
