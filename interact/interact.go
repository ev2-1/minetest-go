package interact

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		m, ok := pkt.Cmd.(*mt.ToSrvInteract)

		if ok {
			go interact(c, m)
		}
	})
}

func interact(c *minetest.Client, m *mt.ToSrvInteract) {
	switch thing := m.Pointed.(type) {
	case *mt.PointedNode:
		pos := thing.Under

		switch m.Action {
		case mt.Dig:
		case mt.Dug:
			Dig(pos)

		case mt.Place:
			Place(c, m)
		}

	default:
		return
	}
}
