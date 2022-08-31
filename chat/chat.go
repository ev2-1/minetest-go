package chat

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"strings"
)

var logger *log.Logger

const loggingSuffix = "[CHAT] "

// Log loggs like its a chat msg
func Log(str string) {
	log.Println(loggingSuffix + str)
}

func init() {
	minetest.RegisterStage1(func() {
		logger = log.New(log.Writer(), "", log.Flags())
	})

	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		cmd, ok := pkt.Cmd.(*mt.ToSrvChatMsg)
		if ok {
			// check if has prefix:
			if strings.HasPrefix(cmd.Msg, cmdPrefix) {
				handleCmd(c, cmd.Msg)
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
