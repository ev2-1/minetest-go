package minetest

import (
	"github.com/ev2-1/minetest-go"
	"github.com/anon55555/mt"

	"reflect"
	"log"
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

	go func() {
		for {
			l, ok := <-minetest.LeaveChan()

			if !ok {
				log.Print("[ERROR]", "Leave channel closed!")
				return
			}

			deleteClt(l.Client)
		}
	}()
}
