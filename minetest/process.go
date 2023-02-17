package minetest

import (
	"github.com/anon55555/mt"

	"fmt"
)

func (c *Client) process(pkt *mt.Pkt) {
	lpkts, ok := GetConfig("log-packets", false)

	if (ConfigVerbose() && !(ok && !lpkts)) || lpkts {
		c.Log("->", fmt.Sprintf("%T", pkt.Cmd))

		defer c.Log("->", fmt.Sprintf("%T done", pkt.Cmd))
	}

	pktProcessorsMu.RLock()
	for _, h := range pktProcessors {
		h(c, pkt)
	}
	pktProcessorsMu.RUnlock()

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
}
