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

	rawPktProcessorsMu.RLock()
	for h := range rawPktProcessors {
		ch := make(chan struct{})
		timeout := makePktTimeout()

		go func(h RawPktProcessor) {
			h(c, pkt)

			close(ch)
		}(h.Thing)

		select {
		case <-ch:
			continue
		case <-timeout.C:
			c.Logf("Timeout waiting for rawPktProcessor! pkt: %T, registerd at %s\n\n", pkt.Cmd, h.Path())
		}
	}
	rawPktProcessorsMu.RUnlock()

	if _, ok := Clts()[c]; !ok && ConfigVerbose() {
		c.Logf("Clt not registerd yet, ignoring for normal pktProcessors")

		// check if invalid pkt type while init seq:
		switch pkt.Cmd.(type) {
		//Allowed packets:
		case *mt.ToSrvInit, *mt.ToSrvFirstSRP, *mt.ToSrvInit2, *mt.ToSrvCltReady:
			break

		default:
			c.Logf("Clt used unexpected Packet while not registerd: %T\n", pkt.Cmd)
			c.SendCmd(&mt.ToCltKick{
				Reason: mt.UnexpectedData,
			})

			//!!ABORT!!
			c.Peer.Conn.Close()
		}

		return
	}

	pktProcessorsMu.RLock()
	for h := range pktProcessors {
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
			c.Logf("Timeout waiting for pktProcessor! pkt: %T, registerd at %s\n\n", pkt.Cmd, h.Path())
		}
	}
	pktProcessorsMu.RUnlock()
}
