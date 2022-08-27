package chat

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
)

// sets the command prefix so `str`
// will be "/" if not set
func SetPrefix(str string) {
	cmdPrefix = str
}

func SendMsg(c *minetest.Client, msg string) {
	c.SendCmd(&mt.ToCltChatMsg{
		Type: mt.RawMsg,

		Text: msg,
	})
}

func SendMsgf(c *minetest.Client, format string, a ...any) {
	SendMsg(c, fmt.Sprintf(format, a...))
}
