package minetest

import (
	"log"
	"plugin"

	"github.com/anon55555/mt"
)

func pluginHook(pl []*plugin.Plugin) {
	for _, p := range pl {
		l, err := p.Lookup("ProcessPkt")

		if err == nil {
			f, ok := l.(func(*Client, *mt.Pkt))
			if !ok {
				log.Println("plugin has wrong 'ProcessPkt' type please check")
				return
			}

			pktProcessors = append(pktProcessors, f)
		}
	}
}
