package minetest

import (
	"github.com/anon55555/mt"
	//	"github.com/anon55555/mt/rudp"

	"reflect"
	//	"strings"
	//	"net"
	"fmt"
)

func (c *Client) process(pkt *mt.Pkt) {
	t := reflect.TypeOf(pkt.Cmd)

	c.Log("->", fmt.Sprintf("%T", pkt.Cmd), t)

	var handled bool

	for _, h := range packetHandlers {
		if h.Packets[t] && h.Handle != nil {
			if h.Handle(c, pkt) {
				handled = true
			}
		}
	}

	if handled {
		return
	}

	switch pkt.Cmd.(type) {
	case *mt.ToSrvCltReady:
		c.SetState(CsActive)
		close(c.initCh)
		return

	}
}
