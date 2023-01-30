package mapLoader

import (
	"github.com/ev2-1/minetest-go/minetest"
	tp "github.com/ev2-1/minetest-go/tools/pos"

	"github.com/anon55555/mt"

	"sync"
	"time"
)

var (
	lastPos   = map[*minetest.Client][3]int16{}
	lastPosMu sync.RWMutex
)

func init() {
	tp.RegisterPosUpdater(func(c *minetest.Client, pos *tp.ClientPos, lu time.Duration) {
		lastPosMu.Lock()
		defer lastPosMu.Unlock()

		apos := pos.Pos.Pos().Int()
		c.Logf("pos: (%5d %5d %5d)", apos[0], apos[1], apos[2])
		ip, _ := mt.Pos2Blkpos(apos)

		p, ok := lastPos[c]
		if ok {
			if p == ip {
				return
			}
		}

		c.Logf("blkpos changed! (%5d %5d %5d)", ip[0], ip[1], ip[2])

		go loadAround(ip, c)
		lastPos[c] = ip
	})
}
