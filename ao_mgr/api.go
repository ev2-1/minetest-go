package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"log"
)

// the first one is reserved for the playerAOID
const lowestAOID = mt.AOID(1)

func GetAOID() mt.AOID {
	aosMu.Lock()
	defer aosMu.Unlock()

	for id := GlobalAOIDmin; id < GlobalAOIDmax; id++ {
		if _, ok := aos[id]; !ok {
			aos[id] = Global

			return id
		}
	}

	return 0
}

func FreeAOID(id mt.AOID) {
	aosMu.Lock()
	defer aosMu.Unlock()

	delete(aos, id)
}

func RmAO(ids ...mt.AOID) {
	for _, id := range ids {
		if id == 0 {
			continue
		}

		FreeAOID(id)
		rmQueueMu.Lock()
		rmQueue[id] = struct{}{}
		rmQueueMu.Unlock()
	}
}

func AOMsg(msgs ...mt.IDAOMsg) {
	for _, msg := range msgs {
		if msg.ID == 0 {
			continue
		}

		globalMsgsMu.Lock()
		globalMsgs = append(globalMsgs, msg)
		globalMsgsMu.Unlock()
	}
}

// GetAllAOIDs returns a slice of all ActiveObjects' ids
func GetAllAOIDs() []mt.AOID {
	activeObjectsMu.RLock()
	activeObjectsMu.RUnlock()

	s := make([]mt.AOID, len(activeObjects))
	i := 0

	for id := range activeObjects {
		s[i] = id
		i++
	}

	return s
}

func GetCltAOIDs(c *minetest.Client) []mt.AOID {
	clientsMu.RLock()
	defer clientsMu.RUnlock()

	if cd, ok := clients[c]; ok {
		s := make([]mt.AOID, len(cd.aos))
		i := 0

		for id := range cd.aos {
			s[i] = id
			i++
		}

		return s
	}

	return []mt.AOID{}
}

// - abstr -

// RegisterAO registers a initialized ActiveObject
func RegisterAO(ao ActiveObject) mt.AOID {
	return registerAO(ao, TypeNormal)
}

func registerAO(ao ActiveObject, t AOType) mt.AOID {
	if ao.GetID() == 0 {
		ao.SetID(GetAOID())
	}

	activeObjectsMu.Lock()
	activeObjects[ao.GetID()] = ao
	activeObjectsMu.Unlock()

	return ao.GetID()
}

var ao0maker func(clt *minetest.Client, id mt.AOID) ActiveObject

// Register player AO0 / self
// RegisterSelfAOMaker is used to register the AO maker for each client
func RegisterAO0Maker(f func(clt *minetest.Client, id mt.AOID) ActiveObject) {
	if ao0maker == nil {
		ao0maker = f
	} else {
		panic(fmt.Errorf("[ao_mgr] Repeated AO0Maker registration attempt."))
	}
}

func SetCltAOID(clt *minetest.Client, id mt.AOID) {
	clt.SetData("aoid", id)
}

func GetCltAOID(c *minetest.Client) (mt.AOID, bool) {
	dat, ok := c.GetData("aoid")
	if !ok {
		return 0, false
	}

	id, ok := dat.(mt.AOID)
	if ok {
		return id, true
	} else {
		log.Fatalf("ClientData has unexpected Type expected %T got %T!\n", mt.AOID(0), dat)
	}

	return 0, false
}
