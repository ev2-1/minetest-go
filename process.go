package minetest

import (
	"github.com/anon55555/mt"
	//	"github.com/anon55555/mt/rudp"

	"reflect"
	//	"strings"
	//	"net"
	"fmt"
)

var pktProcessors []func(*Client, *mt.Pkt)

func (c *Client) process(pkt *mt.Pkt) {
	t := reflect.TypeOf(pkt.Cmd)

	c.Log("->", fmt.Sprintf("%T", pkt.Cmd), t)

	var handled bool

	for _, h := range pktProcessors {
		h(c, pkt)
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
