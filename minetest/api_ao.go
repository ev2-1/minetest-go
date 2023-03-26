package minetest

import (
	"github.com/anon55555/mt"
)

// RegisterAO registers the ActiveObject
func RegisterAO(ao ActiveObject) mt.AOID {
	id := getAOID()
	ao.SetAO(id)

	registerAO(ao)

	return id
}

// RmAO removes AO after calling Clean on AO
// Returns false when ao was not registerd
func RmAO(id mt.AOID) bool {
	ActiveObjectsMu.Lock()
	defer ActiveObjectsMu.Unlock()

	ao, ok := ActiveObjects[id]
	if !ok {
		Loggers.Warnf("tried to delete unregistered AO (aoid: %d)\n", 1, id)

		return false
	}

	ao.Clean()
	delete(ActiveObjects, id)

	return true
}

func GetAO(id mt.AOID) ActiveObject {
	ActiveObjectsMu.RLock()
	defer ActiveObjectsMu.RUnlock()

	return ActiveObjects[id]
}

func BroadcastAOMsgs(ao ActiveObject, msgs ...mt.AOMsg) (<-chan struct{}, error) {
	var merr = new(MultiError)
	var acks []<-chan struct{}

	for clt := range Clts() {
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

func GetPAOID(clt *Client) mt.AOID {
	return GetPAO(clt).GetAO()
}

// GetPAO returns the ActiveObject representing clt.
func GetPAO(clt *Client) ActiveObjectPlayer {
	cd := clt.AOData

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

func ListAOs() map[mt.AOID]ActiveObject {
	ActiveObjectsMu.RLock()
	defer ActiveObjectsMu.RUnlock()

	m := make(map[mt.AOID]ActiveObject, len(ActiveObjects))

	for k, v := range ActiveObjects {
		m[k] = v
	}

	return m
}

func HasAO(clt *Client, ao mt.AOID) bool {
	cd := clt.AOData
	cd.RLock()
	defer cd.RUnlock()

	_, ok := cd.AOs[ao]
	return ok
}

func IDAOMsgs(id mt.AOID, msgs ...mt.AOMsg) (s []mt.IDAOMsg) {
	s = make([]mt.IDAOMsg, len(msgs))

	for k := range msgs {
		s[k] = mt.IDAOMsg{
			ID:  id,
			Msg: msgs[k],
		}
	}

	return s
}

func PPos2AOPos(ppos PPos) AOPos {
	return AOPos{
		Pos: ppos.Pos.Pos,
		Rot: mt.Vec{1: ppos.Pos.Yaw},

		Dim: ppos.Dim,
	}
}
