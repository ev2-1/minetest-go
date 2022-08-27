package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/tools/pos"

	"sync"
)

var (
	globalMsgsMu sync.RWMutex
	globalMsgs   []mt.IDAOMsg
	cltMsgsMu    sync.RWMutex
	cltMsgs      map[*minetest.Client][]mt.IDAOMsg

	// the global queue for AOs to be deleted
	rmsMu sync.RWMutex
	rms   []mt.AOID
)

func (cd *ClientData) doAddQueue() (a []mt.AOAdd) {
	if len(cd.queueAdd) == 0 {
		return
	}

	activeObjectsMu.RLock()
	for _, id := range cd.queueAdd {
		if id != 0 {
			a = append(a, mt.AOAdd{
				ID:       id,
				InitData: activeObjects[id].InitPkt(id, cd.clt),
			})
		} else {
			a = append(a, mt.AOAdd{
				ID:       0,
				InitData: ao0maker(cd.clt).InitPkt(0, cd.clt),
			})
		}
	}
	activeObjectsMu.RUnlock()

	return
}

func init() {
	minetest.RegisterPktTickHook(func() {
		// adds / rm
		clientsMu.RLock()

		for clt, cd := range clients {
			if !cd.initialized.Load() {
				continue
			}

			add := cd.doAddQueue()
			rm := cd.queueRm

			if len(add) != 0 || len(rm) != 0 {
				ack, err := clt.SendCmd(&mt.ToCltAORmAdd{
					Add:    add,
					Remove: rm,
				})

				if err == nil {
					<-ack

					// clear data & update c.aos (if needed)
					if len(add) != 0 {
						cd.aosMu.Lock()
						for _, msg := range add {
							cd.aos[msg.ID] = struct{}{}
						}
						cd.aosMu.Unlock()

						cd.queueAdd = nil
					}

					if len(rm) != 0 {
						cd.queueRm = nil

						cd.aosMu.Lock()
						for _, id := range rm {
							if _, ok := cd.aos[id]; ok {
								delete(cd.aos, id)
							}
						}
						cd.aosMu.Unlock()
					}
				}
			}
		}

		clientsMu.RUnlock()

		// msgs
		globalMsgsMu.RLock()
		cltMsgsMu.RLock()
		for clt := range minetest.Clts() {
			msgs := FilterRelevantMsgs(pos.GetPos(clt).Pos(), append(globalMsgs, cltMsgs[clt]...))

			if len(msgs) != 0 {
				clt.SendCmd(&mt.ToCltAOMsgs{
					Msgs: msgs,
				})
			}
		}
		globalMsgsMu.RUnlock()
		cltMsgsMu.RUnlock()

		globalMsgsMu.Lock()
		if len(globalMsgs) != 0 {
			globalMsgs = make([]mt.IDAOMsg, 0)
		}
		globalMsgsMu.Unlock()

		cltMsgsMu.Lock()
		if len(cltMsgs) != 0 {
			cltMsgs = make(map[*minetest.Client][]mt.IDAOMsg)
		}
		cltMsgsMu.Unlock()
	})
}
