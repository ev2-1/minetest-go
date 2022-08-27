package chat

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/mattn/go-shellwords"

	"fmt"
	"log"
	"strings"
	"sync"
)

var cmdPrefix = "/"

func handleCmd(c *minetest.Client, msg string) {
	msg = strings.TrimPrefix(msg, cmdPrefix)

	args, err := shellwords.Parse(msg)
	if err != nil {
		log.Println("[cmd] error parsing message ", msg, "error: ", err)
		return
	}

	if len(args) == 0 {
		log.Println("[cmd] error: no arguments")
		return
	}

	cmd := args[0]
	args = args[1:]

	cmdsMu.RLock()
	defer cmdsMu.RUnlock()

	h, ok := cmds[cmd]
	if ok {
		h(c, args)
	} else {
		c.SendCmd(&mt.ToCltChatMsg{
			Type: mt.SysMsg,

			Text: "Invalid Comand",
		})
	}
}

var (
	cmds   = make(map[string]func(c *minetest.Client, args []string))
	cmdsMu sync.RWMutex
)

func RegisterChatCmd(name string, f func(c *minetest.Client, args []string)) {
	cmdsMu.Lock()
	defer cmdsMu.Unlock()

	if _, ok := cmds[name]; ok {
		log.Println(fmt.Sprintf("[warning] [cmd] overwriting command \"%s\"", name))
	}

	cmds[name] = f
}
