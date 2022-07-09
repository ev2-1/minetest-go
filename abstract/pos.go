package minetest

import (
	"github.com/ev2-1/minetest-go"
	"github.com/anon55555/mt"

	"sync"
	"time"
)

type CltPos struct {
	*minetest.Client
	*mt.PlayerPos

	LastUpdate int64
}

var posCh = make(chan *CltPos)
var pos   = make(map[*minetest.Client]*mt.PlayerPos)
var posMu sync.RWMutex

var posUpdate   = make(map[*minetest.Client]int64)
var posUpdateMu sync.RWMutex 

// updates pos (TODO :AND does anti cheat)
func updatePos(c *minetest.Client, p *mt.PlayerPos) bool {
	// if posWorks()

	posMu.Lock()
	pos[c] = p
	posCh <- &CltPos{
		Client: c,
		PlayerPos: p,
		
		LastUpdate: time.Now().Unix(),
	}
	posMu.Unlock()

	posUpdateMu.Lock()
	posUpdate[c] = time.Now().Unix()
	posUpdateMu.Unlock()
	
	return false
}

// GetPos returns pos os player / client
func GetPos(c *minetest.Client) mt.PlayerPos {
	posMu.RLock()
	defer posMu.RUnlock()

	return *pos[c]
}

// GetPosChan returns a channel where the updated pos of clts will be sent
func GetPosCh() <-chan *CltPos {
	return posCh
}

// deleteClt
func deleteClt(c *minetest.Client) {
	posMu.Lock()
	defer posMu.Unlock()

	delete(pos, c)
}
