package minetest

import (
	"log"
	"plugin"

	"github.com/anon55555/mt"
)

var leaveHooks []func(*Leave)
var joinHooks []func(*Client)
var initHooks []func(*Client)
var tickHooks []func()
var pktTickHooks []func()
var packetPre []func(*Client, mt.Cmd) bool

func pluginHook(pl map[string]*plugin.Plugin) {
	for _, p := range pl { // no need for Mutexes as are only written once at startup
		l, err := p.Lookup("ProcessPkt")

		if err == nil {
			f, ok := l.(func(*Client, *mt.Pkt))
			if !ok {
				log.Println("plugin has wrong 'ProcessPkt' type please check")
				return
			}

			pktProcessors = append(pktProcessors, f)
		}

		// LeaveHooks:
		l, err = p.Lookup("LeaveHook")
		if err == nil {
			f, ok := l.(func(*Leave))
			if !ok {
				log.Println("plugin has wrong 'LeaveHook' type please check")
				return
			}

			leaveHooks = append(leaveHooks, f)
		}

		// JoinHooks:
		l, err = p.Lookup("JoinHook")
		if err == nil {
			f, ok := l.(func(*Client))
			if !ok {
				log.Println("plugin has wrong 'JoinHooks' type please check")
				return
			}

			joinHooks = append(joinHooks, f)
		}

		// InitHooks:
		l, err = p.Lookup("InitHook")
		if err == nil {
			f, ok := l.(func(*Client))
			if !ok {
				log.Println("plugin has wrong 'InitHook' type please check")
				return
			}

			initHooks = append(initHooks, f)
		}

		// func(*Client, mt.Pkt)
		l, err = p.Lookup("PacketPre")
		if err == nil {
			f, ok := l.(func(*Client, mt.Cmd) bool)
			if !ok {
				log.Println("plugin has wrong 'PacketPre' type please check")
				return
			}

			packetPre = append(packetPre, f)
		}

		// func()
		l, err = p.Lookup("Tick")
		if err == nil {
			f, ok := l.(func())
			if !ok {
				log.Println("plugin has wrong 'Tick' type please check")
				return
			}

			tickHooks = append(tickHooks, f)
		}

		// func()
		l, err = p.Lookup("PktTick")
		if err == nil {
			f, ok := l.(func())
			if !ok {
				log.Println("plugin has wrong 'PktTick' type please check")
				return
			}

			pktTickHooks = append(pktTickHooks, f)
		}
	}
}
