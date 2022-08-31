package minetest_map

import (
	"github.com/EliasFleckenstein03/mtmap"
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	toolspos "github.com/ev2-1/minetest-go/tools/pos"

	"sync"
	"time"
)

// a list of all clients and their loaded chunks
var loadedChunks = make(map[*minetest.Client]map[pos]bool)
var loadedChunksMu sync.RWMutex

var (
	MapBlkUpdateRate, _ = time.ParseDuration("2s") // in seconds

	MapBlkUpdateRange  = int16(10) // in mapblks
	MapBlkUpdateHeight = int16(5)  // in mapblks

	heigthOff = -MapBlkUpdateHeight / 2
)

var stone mt.Content
var grass mt.Content
var exampleBlk mtmap.MapBlk

func init() {
	OpenDB(minetest.Path("/map.sqlite"))

	minetest.RegisterStage2(stage2)
}

func stage2() {
	s := minetest.GetNodeDef("mcl_core:stone")
	if s != nil {
		stone = s.Param0
	}

	s = minetest.GetNodeDef("mcl_core:dirt_with_grass")
	if s != nil {
		grass = s.Param0
	}

	exampleBlk = mtmap.MapBlk{}

	for i := 0; i < 4096; i++ {
		exampleBlk.Param0[i] = 126
	}

	for i := 0; i < 16*16; i++ {
		exampleBlk.Param0[i] = stone
	}

	// center block is stone:
	exampleBlk.Param0[4096/2+16/2] = grass // some wool
}

func init() {
	toolspos.RegisterPosUpdater(func(c *minetest.Client, pos mt.PlayerPos, LastUpdate int64) {
		p := Pos2int(pos.Pos())
		blkpos, _ := mt.Pos2Blkpos(p)

		go func() {
			for _, sp := range spiral(MapBlkUpdateRange) {
				for i := int16(0); i < MapBlkUpdateRange; i++ {
					// generate absolute position
					ap := sp.add(blkpos).add([3]int16{0, heigthOff + i})

					// load block
					LoadBlk(c, ap, false)
				}
			}
		}()
	})
}
