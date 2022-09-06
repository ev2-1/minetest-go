package mapLoader

import (
	"github.com/ev2-1/minetest-go/minetest"
	tp "github.com/ev2-1/minetest-go/tools/pos"

	"github.com/anon55555/mt"

	"sync"
)

var (
	lastPos   = map[*minetest.Client][3]int16{}
	lastPosMu sync.RWMutex
)

func init() {
	tp.RegisterPosUpdater(func(c *minetest.Client, pos mt.PlayerPos, lu int64) {
		lastPosMu.Lock()
		defer lastPosMu.Unlock()

		ip, _ := mt.Pos2Blkpos(pos.Pos().Int())

		p, ok := lastPos[c]
		if ok {
			if p == ip {
				return
			}
		}

		c.Log("blkpos changed! (", ip[0], ip[1], ip[2], ")")

		go loadAround(ip, c)
		lastPos[c] = ip
	})
}
