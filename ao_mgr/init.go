package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"sync"
)

// Data kept per client
type ClientData struct {
	clt *minetest.Client

	initialized bool // player AO initialized; wont send anything until true (gets set true when `queueAddSelf` is invoked)

	// which AOs do you have?
	aos map[mt.AOID]struct{}

	// the id you have yourself
	id mt.AOID

	// queues
	queueAdd []mt.AOID
	queueRm  []mt.AOID
}

func (cd *ClientData) QueueAdd(adds ...mt.AOID) {
	cd.clt.Log(fmt.Sprintf("Adding AOIDs %v to client add queue", adds))

	cd.queueAdd = append(cd.queueAdd, adds...)

	for _, id := range adds {
		cd.aos[id] = struct{}{}
	}
}

func makeClientData(c *minetest.Client) *ClientData {
	return &ClientData{
		clt: c,

		aos: make(map[mt.AOID]struct{}),
	}
}

var clients = make(map[*minetest.Client]*ClientData)
var clientsMu sync.RWMutex

func init() {
	minetest.RegisterJoinHook(func(clt *minetest.Client) {
		cd := makeClientData(clt)

		// give client data
		clientsMu.Lock()
		clients[clt] = cd
		clientsMu.Unlock()

		if ao0maker == nil {
			panic(fmt.Errorf("no AO0Maker registerd, please ensure you have a player managing plugin installed."))
		}

		// make playerAO
		ao := ao0maker(clt)
		id := RegisterAO(ao)

		// forceignore id for self:
		cd.aos[id] = struct{}{}
		cd.id = id

		// add self to schedule first:
		go func(id mt.AOID, cd *ClientData) {
			ack, _ := clt.SendCmd(&mt.ToCltAORmAdd{
				Add: []mt.AOAdd{
					mt.AOAdd{
						ID:       0,
						InitData: ao.InitPkt(0, cd.clt),
					},
				},
			})

			<-ack

			cd.initialized = true
		}(id, cd)
	})

	minetest.RegisterLeaveHook(func(l *minetest.Leave) {
		clientsMu.RLock()
		defer clientsMu.RUnlock()
		if cd, ok := clients[l.Client]; ok {
			RmAO(cd.id)

			delete(clients, l.Client)
		}
	})
}
