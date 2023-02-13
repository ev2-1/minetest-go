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

		for clt, d := range clients {
			d.aosMu.Lock()

			for id, ao := range activeObjects {
				if t, ok := d.aos[id]; !ok {
					if t != TypeNormal {
						continue
					}

					// clt dosn't have AO, check if relevant:
					if RelevantAO(clt, ao) {
						d.QueueAdd(id)
						clt.Logf("adding AO %d\n", id)
					}
				}
			}

			d.aosMu.Unlock()
		}

		rmQueueMu.Lock()

		// remove
		for clt, d := range clients {
			d.aosMu.RLock()

			for id, t := range d.aos {
				cid, ok := GetCltAOID(clt)

				// Ignore special AOs & self
				if t != TypeNormal || (ok && cid == id) {
					continue
				}

				ao := activeObjects[id]

				// should be removed        || out of range
				if _, ok := rmQueue[id]; ok || !RelevantAO(clt, ao) {
					// clt has stuff from rmqueue:
					d.QueueRm(id)
					clt.Logf("removing AO %d\n", id)
				}
			}

			d.aosMu.RUnlock()
		}

		clientsMu.RUnlock()
		activeObjectsMu.RUnlock()

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
