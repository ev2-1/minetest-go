package main

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"

	"log"
	"time"
)

func ProcessPkt(c *minetest.Client, pkt *mt.Pkt) {
	switch cmd := pkt.Cmd.(type) {
	case *mt.ToSrvChatMsg:
		log.Printf("[CHAT] <%s> %s", c.Name, cmd.Msg)

		ts := time.Now().Unix()

		go minetest.Broadcast(&mt.ToCltChatMsg{
			Sender: c.Name,
			Text:   cmd.Msg,

			Type: mt.NormalMsg,

			Timestamp: ts,
		})
	}

	return
}
