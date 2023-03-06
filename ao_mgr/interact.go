package ao

import (
	"github.com/ev2-1/minetest-go/minetest"
	//	"github.com/ev2-1/minetest-go/minetest/log"

	"github.com/anon55555/mt"
)

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		i, ok := pkt.Cmd.(*mt.ToSrvInteract)
		if !ok {
			return
		}

		pt, ok := i.Pointed.(*mt.PointedAO)
		if !ok {
			return
		}

		interact(c, i, pt)
	})
}

func interact(c *minetest.Client, i *mt.ToSrvInteract, ao *mt.PointedAO) {
	if i.Action != mt.Dig { // unexpected
		c.Logf("Unexpected interaction with AO(%d): %s\n", ao.ID, i.Action)

		return
	}

	_ao := GetAO(ao.ID)

	// get corresponding AO
	// check if client has AO:
	if !HasAO(c, ao.ID) {
		c.Logf("Unexpeted interaction with AO(%d) client does not have AO! (ao is type %T)", ao.ID, _ao)

		return
	}

	if _ao == nil {
		c.Logf("Client either had <nil> AO or interacted with invalid AO")

		return
	}

	_ao.Punch(c, i)
}

func HasAO(clt *minetest.Client, ao mt.AOID) bool {
	cd := GetClientData(clt)
	cd.RLock()
	defer cd.RUnlock()

	_, ok := cd.AOs[ao]
	return ok
}
