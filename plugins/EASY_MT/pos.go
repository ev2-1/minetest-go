package main

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"

	"log"
	"plugin"
	"sync"
	"time"
)

var pos = make(map[*minetest.Client]*mt.PlayerPos)
var posMu sync.RWMutex

var posUpdate = make(map[*minetest.Client]int64)
var posUpdateMu sync.RWMutex

func PluginsLoaded(pl map[string]*plugin.Plugin) {
	for _, p := range pl {
		l, err := p.Lookup("PosUpdate")

		if err == nil {
			f, ok := l.(func(*minetest.Client, *mt.PlayerPos, int64))
			if !ok {
				log.Println("[EASY_MT] PosUpdate callback error, check plugin!")
				return
			}

			posUpdaters = append(posUpdaters, f)
		}
	}
}

var posUpdaters []func(c *minetest.Client, pos *mt.PlayerPos, lu int64)

func updatePos(c *minetest.Client, p *mt.PlayerPos) {
	posUpdateMu.RLock()

	for _, u := range posUpdaters {
		u(c, p, posUpdate[c])
	}

	posUpdateMu.RUnlock()
	posUpdateMu.Lock()

	posUpdate[c] = time.Now().Unix()

	posUpdateMu.Unlock()
}

// GetPos returns pos os player / client
func GetPos(c *minetest.Client) mt.PlayerPos {
	posMu.RLock()
	defer posMu.RUnlock()

	return *pos[c]
}

// deleteClt
func deleteClt(c *minetest.Client) {
	posMu.Lock()
	defer posMu.Unlock()

	delete(pos, c)
}
