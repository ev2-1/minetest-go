package main

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"

	"reflect"
)

func initInteractions() {
	minetest.RegisterPacketHandler(&minetest.PacketHandler{
		Packets: map[reflect.Type]bool{
			reflect.TypeOf(&mt.ToSrvInteract{}): true,
		},

		Handle: func(c *minetest.Client, pkt *mt.Pkt) bool {
			switch cmd := pkt.Cmd.(type) {
			case *mt.ToSrvInteract:
				interact(cmd)
			}

			return false
		},
	})
}

func interact(m *mt.ToSrvInteract) {
	switch thing := m.Pointed.(type) {
	case *mt.PointedNode:
		pos := thing.Under

		switch m.Action {
		case mt.Dig:
		case mt.Dug:
			SetNode(pos, mt.Air)
		}

	default:
		return
	}
}
