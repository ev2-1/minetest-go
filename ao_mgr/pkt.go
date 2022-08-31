package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	//	"github.com/ev2-1/minetest-go/tools/pos"

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
		if id != 0 && activeObjects[id] != nil {
			a = append(a, mt.AOAdd{
				ID:       id,
				InitData: activeObjects[id].InitPkt(cd.clt),
			})
		} else {
			a = append(a, mt.AOAdd{
				ID:       0,
				InitData: ao0maker(cd.clt, 0).InitPkt(cd.clt),
			})
		}
	}
	activeObjectsMu.RUnlock()

	return
}

func init() {
	minetest.RegisterPktTickHook(func() {
		activeObjectsMu.RLock()
		defer activeObjectsMu.RUnlock()

		for _, ao := range activeObjects {
			msgs, f := ao.Pkts()
			if f {
				id := ao.GetID()

				idmsgs := make([]mt.IDAOMsg, len(msgs))

				for i := 0; i < len(msgs); i++ {
					idmsgs[i] = mt.IDAOMsg{
						ID:  id,
						Msg: msgs[i],
					}
				}

				AOMsg(idmsgs...)
			}
		}
	})

	minetest.RegisterPktTickHook(func() {
		// adds / rm
		clientsMu.RLock()

		for clt, cd := range clients {
			if !cd.initialized.Load() {
				clt.Log("not yet initialized")

				continue
			}

			add := cd.doAddQueue()
			rm := cd.queueRm

			if len(add) != 0 || len(rm) != 0 {
				ack, err := clt.SendCmd(&mt.ToCltAORmAdd{
					Add:    add,
					Remove: rm,
				})

				if err != nil {
					clt.Log("error sending AOS, retrying next tick")
				}

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

		clientsMu.RUnlock()

		// msgs
		globalMsgsMu.Lock()
		cltMsgsMu.RLock()
		for clt := range minetest.Clts() {
			//			msgs := FilterRelevantMsgs(pos.GetPos(clt).Pos(), append(globalMsgs, cltMsgs[clt]...))
			msgs := globalMsgs
			if len(msgs) != 0 {
				clt.SendCmd(&mt.ToCltAOMsgs{
					Msgs: msgs,
				})
			}
		}
		cltMsgsMu.RUnlock()

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
