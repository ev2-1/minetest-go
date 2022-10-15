package minetest_map

import (
	"github.com/EliasFleckenstein03/mtmap"
	"github.com/ev2-1/minetest-go/minetest"

	"sync"
)

// SBM describes a Send Block Manager
// gets called whenever a mapblk is send to a client
type SBM struct {
	Send func(clt *minetest.Client, pos [3]int16, blk *mtmap.MapBlk)
}

var _SBMs []*SBM
var _SMBsMu sync.RWMutex

func doSBM(c *minetest.Client, p [3]int16, blkdata *mtmap.MapBlk) {
	_SMBsMu.RLock()
	defer _SMBsMu.RUnlock()

	for _, s := range _SBMs {
		if s.Send != nil {
			s.Send(c, p, blkdata)
		}
	}
}

// RegisterSBM registers a SBM
func RegisterSBM(s *SBM) {
	_SMBsMu.Lock()
	defer _SMBsMu.Unlock()

	_SBMs = append(_SBMs, s)
}
