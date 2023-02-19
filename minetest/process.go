package minetest

import (
	"github.com/anon55555/mt"

	"fmt"
	"time"
)

const defaultDuration = "10s"

func makePktTimeout() *time.Timer {
	dstr := GetConfigV("pkt-timeout", defaultDuration)

	duration, err := time.ParseDuration(dstr)
	if err != nil {
		duration, err = time.ParseDuration(defaultDuration)
		if err != nil {
			panic(err)
		}
	}

	return time.NewTimer(duration)
}

func (c *Client) process(pkt *mt.Pkt) {
	lpkts, ok := GetConfig("log-packets", false)

	if (ConfigVerbose() && !(ok && !lpkts)) || lpkts {
		c.Log("->", fmt.Sprintf("%T", pkt.Cmd))

		defer c.Log("->", fmt.Sprintf("%T done", pkt.Cmd))
	}

	pktProcessorsMu.RLock()
	for _, h := range pktProcessors {
		ch := make(chan struct{})
		timeout := makePktTimeout()

		go func(h PktProcessor) {
			h(c, pkt)

			close(ch)
		}(h.Thing)

		select {
		case <-ch:
			continue
		case <-timeout.C:
			c.Logf("Timeout waiting for pktProcessor! pkt: %T, registerd at %s\n", pkt.Cmd, h.Path)
		}
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
