package main

import (
	"github.com/ev2-1/minetest-go"
	"github.com/anon55555/mt"

	"reflect"
	"time"
	"log"
)

func init() {
	minetest.RegisterPacketHandler(&minetest.PacketHandler{
		Packets: map[reflect.Type]bool{
			reflect.TypeOf(&mt.ToSrvChatMsg{}): true,
		},

		Handle: func(c *minetest.Client, pkt *mt.Pkt) bool {
			switch cmd := pkt.Cmd.(type) {
			case *mt.ToSrvChatMsg:
				log.Printf("[CHAT] <%s> %s", c.Name, cmd.Msg)

				ts := time.Now().Unix()

				go minetest.Broadcast(&mt.ToCltChatMsg{
					Sender: c.Name,
					Text: cmd.Msg,

					Type: mt.NormalMsg,

					Timestamp: ts,
				})
			}

			return true
		},
	})
}
