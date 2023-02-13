package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"sync"
)

// Data kept per client
type ClientData struct {
	sync.RWMutex
	clt *minetest.Client

	initialized bool // player AO initialized; wont send anything until true (gets set true when `queueAddSelf` is invoked)

	// which AOs do you have?
	aosMu sync.RWMutex
	aos   map[mt.AOID]AOType

	// the id you have yourself
	id mt.AOID

	// queues
	queueAdd map[mt.AOID]struct{}
	queueRm  map[mt.AOID]struct{}
}

//go:generate stringer -type AOType
type AOType uint8

const (
	TypeNormal AOType = iota
	TypeForced        // ignores ao for automatic actions
)

func (cd *ClientData) GetID() mt.AOID {
	cd.RLock()
	defer cd.RUnlock()

	return cd.id
}

func (cd *ClientData) QueueAdd(adds ...mt.AOID) {
	cd.Lock()
	defer cd.Unlock()

	cd.clt.Log(fmt.Sprintf("Adding AOIDs %v to client add queue", adds))

	if cd.queueAdd == nil {
		cd.queueAdd = make(map[mt.AOID]struct{})
	}

	for _, id := range adds {
		cd.queueAdd[id] = struct{}{}
	}
}

func (cd *ClientData) queueRmS() (s []mt.AOID) {
	s = make([]mt.AOID, 0, len(cd.queueRm))

	for k := range cd.queueRm {
		s = append(s, k)
	}

	return
}

func (cd *ClientData) queueAddS() (s []mt.AOID) {
	s = make([]mt.AOID, 0, len(cd.queueAdd))

	for k := range cd.queueAdd {
		s = append(s, k)
	}

	return
}

func (cd *ClientData) QueueRm(rms ...mt.AOID) {
	cd.Lock()
	defer cd.Unlock()

	cd.clt.Log(fmt.Sprintf("Adding AOIDs %v to client rm queue", rms))

	if cd.queueRm == nil {
		cd.queueRm = make(map[mt.AOID]struct{})
	}

	for _, id := range rms {
		cd.queueRm[id] = struct{}{}
	}
}

func makeClientData(c *minetest.Client) *ClientData {
	return &ClientData{
		clt: c,

		aos: make(map[mt.AOID]AOType),

		queueAdd: make(map[mt.AOID]struct{}),
		queueRm:  make(map[mt.AOID]struct{}),
	}
}

var clients = make(map[*minetest.Client]*ClientData)
var clientsMu sync.RWMutex

func init() {
	minetest.RegisterJoinHook(func(clt *minetest.Client) {
		go func() {
			cd := makeClientData(clt)
			cd.Lock()
			defer cd.Unlock()

			// give client data
			clientsMu.Lock()
			clients[clt] = cd
			clientsMu.Unlock()

			if ao0maker == nil {
				panic(fmt.Errorf("no AO0Maker registerd, please ensure you have a player managing plugin installed."))
			}

			// make playerAO (for the others)
			id := GetAOID()
			ao := ao0maker(clt, id)
			registerAO(ao, TypeForced)

			// forceignore id for self:
			cd.aosMu.Lock()
			cd.aos[id] = TypeNormal
			cd.aosMu.Unlock()

			// make playerAO for self:
			ao = ao0maker(clt, 0)

			// add self to schedule first:
			ack, _ := clt.SendCmd(&mt.ToCltAORmAdd{
				Add: []mt.AOAdd{
					mt.AOAdd{
						ID:       0,
						InitData: ao.InitPkt(cd.clt),
					},
				},
			})

			<-ack

			cd.clt.Log("initialized!")
			cd.initialized = true
		}()
	})

	minetest.RegisterLeaveHook(func(l *minetest.Leave) {
		clientsMu.Lock()
		defer clientsMu.Unlock()

		if cd, ok := clients[l.Client]; ok {
			cd.Lock()
			defer cd.Unlock()

			l.Client.Logger.Printf("Removing AO (%d)\n", cd.id)

			RmAO(cd.id)
			delete(clients, l.Client)
		}
	})
}
