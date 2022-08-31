package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"
)

func init() {
	chat.RegisterChatCmd("list_aos", func(c *minetest.Client, args []string) {
		if len(args) == 0 {
			chat.SendMsg(c, "Usage: list_aos <global | local>", mt.SysMsg)
			return
		}

		switch args[0] {
		case "global":
			chat.SendMsgf(c, mt.SysMsg, "Global ActiveObjects (ids): %v", GetAllAOIDs())
			return

		case "local":
			chat.SendMsgf(c, mt.SysMsg, "Your (%s), ActiveObjects (ids): %v", c.Name, GetCltAOIDs(c))
			return
		}
	})
}
