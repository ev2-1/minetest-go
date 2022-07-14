package main

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"

	"reflect"
)

func init() {
	minetest.RegisterPacketHandler(&minetest.PacketHandler{
		Packets: map[reflect.Type]bool{
			reflect.TypeOf(&mt.ToSrvPlayerPos{}): true,
		},

		Handle: func(c *minetest.Client, pkt *mt.Pkt) bool {
			switch cmd := pkt.Cmd.(type) {
			case *mt.ToSrvPlayerPos:
				updatePos(c, &cmd.Pos)
			}

			return false
		},
	})
}
