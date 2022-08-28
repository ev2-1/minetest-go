package minetest_map

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		m, ok := pkt.Cmd.(*mt.ToSrvInteract)

		if ok {
			interact(m)
		}
	})
}

func interact(m *mt.ToSrvInteract) {
	switch thing := m.Pointed.(type) {
	case *mt.PointedNode:
		pos := thing.Under

		switch m.Action {
		case mt.Dig:
		case mt.Dug:
			SetNode(pos, mt.Node{Param0: mt.Air})
		}

	default:
		return
	}
}
