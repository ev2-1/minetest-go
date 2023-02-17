package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"time"
)

const AOActionTimeout time.Duration = time.Second

func init() {
	minetest.RegisterTickHook(func() {
		clts := minetest.Clts()
		startTime := time.Now()

		ActiveObjectsMu.RLock()
		defer ActiveObjectsMu.RUnlock()

		for clt := range clts {
			go func(clt *minetest.Client) {
				cd := GetClientData(clt)
				if cd == nil {
					return
				}

				addQueue := make(map[mt.AOID]*AOInit)
				rmQueue := make(map[mt.AOID]struct{})

				cd.RLock()
				defer cd.RUnlock()

				if !cd.Ready {
					return
				}

				//Look for globally added:
				for id, ao := range ActiveObjects {
					if _, ok := cd.AOs[id]; !ok && id != cd.AOID && Relevant(ao, clt) {
						clt.Logf("scheduling %d for aoadd\n", id)
						addQueue[id] = ao.AOInit(clt)
					}
				}

				//Look for globally removed:
				for id := range cd.AOs {
					if _, ok := ActiveObjects[id]; !ok && id != cd.AOID {
						clt.Logf("scheduling %d for aorm\n", id)
						rmQueue[id] = struct{}{}
					}
				}

				laq := len(addQueue)

				//skip
				if laq <= 0 && len(rmQueue) <= 0 {
					return
				}

				adds := make([]mt.AOAdd, laq)

				if laq > 0 {
					var i int
					for id, init := range addQueue {
						adds[i] = mt.AOAdd{
							ID:       id,
							InitData: init.AOInitData(id),
						}

						i++
					}
				}

				ack, err := clt.SendCmd(&mt.ToCltAORmAdd{
					Remove: map2slice(rmQueue),
					Add:    adds,
				})

				if err != nil {
					clt.Logf("[WARN] Error encounterd when sending pkt: %s\n", err)
					return
				}

				timeout := time.After(AOActionTimeout)

				select {
				case <-timeout:
					clt.Logf("[WARN] AOAction timed out after %s\n", AOActionTimeout)
					return

				case <-ack:
					//TODO: check which has prio in MT code
					// apply to ClientData
					for id := range addQueue {
						cd.AOs[id] = struct{}{}
					}

					for id := range rmQueue {
						delete(cd.AOs, id)
					}
				}

			}(clt)
		}

		duration := time.Now().Sub(startTime)
		if duration > time.Millisecond*64 {
			log.Printf("[WARN] AO tick took %s!\n", duration)
		}
	})
}
