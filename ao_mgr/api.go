package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"sync"
)

const (
	LowestAOID  mt.AOID = (1)
	HighestAOID mt.AOID = (65534) // one lower than largest value
)

var (
	ActiveObjectsMu sync.RWMutex
	ActiveObjects   = make(map[mt.AOID]ActiveObject)
)

func getAOID() mt.AOID {
	ActiveObjectsMu.RLock()
	defer ActiveObjectsMu.RUnlock()

	for id := LowestAOID; id < HighestAOID; id++ {
		if _, ok := ActiveObjects[id]; !ok {
			return id
		}
	}

	panic("No free AOIDs left!")
}

// RegisterAO registers the ActiveObject
func RegisterAO(ao ActiveObject) mt.AOID {
	id := getAOID()
	ao.SetAO(id)

	registerAO(ao)

	return id
}

func registerAO(ao ActiveObject) {
	ActiveObjectsMu.Lock()
	defer ActiveObjectsMu.Unlock()

	ActiveObjects[ao.GetAO()] = ao
}

// RmAO removes AO after calling Clean on AO
// Returns false when ao was not registerd
func RmAO(id mt.AOID) bool {
	ActiveObjectsMu.Lock()
	defer ActiveObjectsMu.Unlock()

	ao, ok := ActiveObjects[id]
	if !ok {
		log.Printf("[WARN] tried to delete unregistered AO (aoid: %d)\n", id)

		return false
	}

	ao.Clean()
	delete(ActiveObjects, id)

	return true
}

const ClientDataKey = "ao_data"

// GetClientData returns a pointer to Clients ClientData
// Nil if not defined
func GetClientData(clt *minetest.Client) *ClientData {
	data, ok := clt.GetData(ClientDataKey)
	if !ok {
		clt.Logf("[WARN] Client does not have ao.ClientData!")
		cd := makeClientData()
		clt.SetData(ClientDataKey, cd)

		return cd
	}

	cd, ok := data.(*ClientData)
	if !ok {
		clt.Fatalf("ClientData at '%s' is not of type '*ao.ClientData' but '%T'", ClientDataKey, data)

		return nil
	}

	return cd
}

func GetAO(id mt.AOID) ActiveObject {
	ActiveObjectsMu.RLock()
	defer ActiveObjectsMu.RUnlock()

	return ActiveObjects[id]
}

func BroadcastAOMsgs(ao ActiveObject, msgs ...mt.AOMsg) (<-chan struct{}, error) {
	var merr = new(MultiError)
	var acks []<-chan struct{}

	for clt := range minetest.Clts() {
		if Relevant(ao, clt) {
			ack, err := clt.SendCmd(&mt.ToCltAOMsgs{
				Msgs: IDAOMsgs(ao.GetAO(), msgs...),
			})

			acks = append(acks, ack)

			if err != nil {
				merr.Add(err)
			}
		}
	}

	if len(merr.Errs) > 0 {
		return Acks(acks...), merr
	} else {
		return Acks(acks...), nil
	}
}

// GetPAO returns the ActiveObject representing clt.
func GetPAO(clt *minetest.Client) ActiveObjectPlayer {
	cd := GetClientData(clt)

	cd.RLock()
	ao := GetAO(cd.AOID)
	cd.RUnlock()

	cdao, ok := ao.(ActiveObjectPlayer)
	if !ok {
		clt.Fatalf("Client AO is not ActiveObjectPlayer but %T!\n", ao)
		return nil
	}

	return cdao
}
