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

	var handled bool

	for _, h := range pktProcessors {
		h(c, pkt)
	}

	if handled {
		return
	}

	switch pkt.Cmd.(type) {
	case *mt.ToSrvCltReady:
		if c.State == CsActive {
			registerPlayer(c)
		} else {
			CltLeave(&Leave{
				Reason: mt.UnexpectedData,

				Client: c,
			})
		}

		close(c.initCh)
		return

	case *mt.ToSrvGotBlks:
		return
	}

	if verbose {
		c.Log("->", fmt.Sprintf("%T", pkt.Cmd), t)
	}

}
