package chat

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"time"
)

// sets the command prefix so `str`
// will be "/" if not set
func SetPrefix(str string) {
	cmdPrefix = str
}

func SendMsg(c *minetest.Client, msg string, cmt mt.ChatMsgType) {
	c.SendCmd(&mt.ToCltChatMsg{
		Type: cmt,

		Text:      msg,
		Timestamp: time.Now().Unix(),
	})
}

func SendMsgf(c *minetest.Client, cmt mt.ChatMsgType, format string, a ...any) {
	SendMsg(c, fmt.Sprintf(format, a...), cmt)
}

func BroadcastMsg(sender, text string) {
	Log(fmt.Sprintf("<%s> %s", sender, text))

	minetest.Broadcast(&mt.ToCltChatMsg{
		Type: mt.NormalMsg,

		Sender:    sender,
		Text:      text,
		Timestamp: time.Now().Unix(),
	})
}

func Broadcast(msg string, cmt mt.ChatMsgType) {
	Log(msg)

	minetest.Broadcast(&mt.ToCltChatMsg{
		Type: cmt,

		Text:      msg,
		Timestamp: time.Now().Unix(),
	})
}

func Broadcastf(cmt mt.ChatMsgType, format string, a ...any) {
	Broadcast(fmt.Sprintf(format, a...), cmt)
}
