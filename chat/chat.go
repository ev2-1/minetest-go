package chat

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"strings"
	"time"
)

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		cmd, ok := pkt.Cmd.(*mt.ToSrvChatMsg)
		if ok {
			// check if has prefix:
			if strings.HasPrefix(cmd.Msg, cmdPrefix) {
				handleCmd(c, cmd.Msg)
			} else {

				log.Printf("[CHAT] <%s> %s", c.Name, cmd.Msg)

				ts := time.Now().Unix()

				go minetest.Broadcast(&mt.ToCltChatMsg{
					Sender: c.Name,
					Text:   cmd.Msg,

					Type: mt.NormalMsg,

					Timestamp: ts,
				})
			}
		}
	})
}
