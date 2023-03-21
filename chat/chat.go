package chat

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/minetest/log"

	"strings"
)

const loggingPrefix = "[CHAT] "

// Log loggs like its a chat msg
func Log(str string) {
	log.Println(loggingPrefix + str)
}

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		cmd, ok := pkt.Cmd.(*mt.ToSrvChatMsg)
		if ok {
			// check if has prefix:
			if strings.HasPrefix(cmd.Msg, cmdPrefix) {
				HandleCmd(c, cmd.Msg)
			} else {
				BroadcastMsg(c.Name, cmd.Msg)
			}
		}
	})

	minetest.RegisterLeaveHook(func(l *minetest.Leave) {
		Broadcastf(mt.SysMsg, "%s left the game.", l.Client.Name)
	})

	minetest.RegisterJoinHook(func(clt *minetest.Client) {
		Broadcastf(mt.SysMsg, "%s joined the game.", clt.Name)
	})
}
