package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
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
		rmQueue = append(rmQueue, id)
	}
}

func AOMsg(msgs ...mt.IDAOMsg) {
	for _, msg := range msgs {
		if msg.ID == 0 {
			continue
		}

		globalMsgsMu.RLock()
		globalMsgs = append(globalMsgs, msg)
		globalMsgsMu.RUnlock()
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
	}

	return []mt.AOID{}
}

// - abstr -

// RegisterAO registers a initialized ActiveObject
func RegisterAO(ao ActiveObject) mt.AOID {
	if ao.GetID() == 0 {
		ao.SetID(GetAOID())
	}

	activeObjectsMu.Lock()
	activeObjects[ao.GetID()] = ao
	activeObjectsMu.Unlock()

	return ao.GetID()
}

var ao0maker func(clt *minetest.Client) ActiveObject

// Register player AO0 / self
// RegisterSelfAOMaker is used to register the AO maker for each client
func RegisterAO0Maker(f func(clt *minetest.Client) ActiveObject) {
	if ao0maker == nil {
		ao0maker = f
	} else {
		panic(fmt.Errorf("[ao_mgr] Repeated AO0Maker registration attempt."))
	}
}
