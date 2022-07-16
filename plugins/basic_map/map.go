package main

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"

	"plugin"
	"time"
)

// a list of all clients and their loaded chunks
var loadedChunks map[string]map[pos]int64

var (
	MapBlkUpdateRate   int64 = 2         // in seconds
	MapBlkUpdateRange        = int16(10) // in mapblks
	MapBlkUpdateHeight       = int16(5)  // in mapblks

	heigthOff = -MapBlkUpdateHeight / 2
	EmptyBlk  mt.MapBlk
)

var stone mt.Content
var grass mt.Content
var exampleBlk mt.MapBlk

func init() {
	loadedChunks = make(map[string]map[pos]int64)
	OpenDB(minetest.Path("/map.sqlite"))
}

func PluginsLoaded(map[string]*plugin.Plugin) {
	s := minetest.GetNodeDef("mcl_core:stone")
	if s != nil {
		stone = s.Param0
	}

	s = minetest.GetNodeDef("mcl_core:dirt_with_grass")
	if s != nil {
		grass = s.Param0
	}

	exampleBlk = mt.MapBlk{}

	for i := 0; i < 4096; i++ {
		exampleBlk.Param0[i] = 126
	}

	for i := 0; i < 16*16; i++ {
		exampleBlk.Param0[i] = stone
	}

	// center block is stone:
	exampleBlk.Param0[4096/2+16/2] = grass // some wool

	for k := range EmptyBlk.Param0 {
		EmptyBlk.Param0[k] = mt.Air
	}
}

func PosUpdate(c *minetest.Client, pos *mt.PlayerPos, LastUpdate int64) {
	if time.Now().Unix() < LastUpdate+MapBlkUpdateRate {
		p := Pos2int(pos.Pos())
		blkpos, _ := mt.Pos2Blkpos(p)

		name := c.Name

		for _, sp := range spiral(MapBlkUpdateRange) {
			for i := int16(0); i < MapBlkUpdateRange; i++ {
				// generate absolute position
				ap := sp.add(blkpos).add([3]int16{0, heigthOff + i})

				// load block
				blk := LoadChunk(name, ap)

				// if block has content; send to clt
				if blk != nil {
					go c.SendCmd(&mt.ToCltBlkData{
						Blkpos: ap,
						Blk:    *blk,
					})
				}
			}
		}
	}
}

func LoadChunk(name string, p pos) *mt.MapBlk {
	if loadedChunks[name] == nil {
		loadedChunks[name] = make(map[pos]int64)
	}

	t := time.Now().Unix()

	if !(loadedChunks[name][p] < t-MapBlkUpdateRate) {
		return nil
	}

	blkdata, ok := GetBlk(p)
	if !ok {
		SetBlk(p, &exampleBlk)
		blkdata = exampleBlk
	}

	loadedChunks[name][p] = t

	return &blkdata
}
