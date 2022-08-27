package ao

import (
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"
)

func init() {
	chat.RegisterChatCmd("list_aos", func(c *minetest.Client, args []string) {
		if len(args) == 0 {
			chat.SendMsg(c, "Usage: list_aos <global | local>")
			return
		}

		switch args[0] {
		case "global":
			chat.SendMsgf(c, "Global ActiveObjects (ids): %v", GetAllAOIDs())
			return

		case "local":
			chat.SendMsgf(c, "Your (%s), ActiveObjects (ids): %v", c.Name, GetCltAOIDs(c))
			return
		}
	})
}
