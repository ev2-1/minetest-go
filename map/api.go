package minetest_map

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"github.com/EliasFleckenstein03/mtmap"
)

// LoadBlk sends a minetest.Client a blk at pos
// TODO: if force is false, will only update every 10 seconds (atm sends only once (per client))
// triggers SBMs
func LoadBlk(c *minetest.Client, p [3]int16, force bool) {
	loadedChunksMu.Lock()
	if loadedChunks[c] == nil {
		loadedChunks[c] = make(map[pos]bool)
	}
	loadedChunksMu.Unlock()

	loadedChunksMu.RLock()
	if !force && loadedChunks[c][p] {
		loadedChunksMu.RUnlock()
		return
	}
	loadedChunksMu.RUnlock()

	ch := GetBlk(p)
	go func() {
		blkdata := <-ch
		if blkdata == nil {
			SetBlk(p, &exampleBlk)
			blkdata = &exampleBlk
		}

		loadedChunksMu.Lock()
		loadedChunks[c][p] = true
		loadedChunksMu.Unlock()

		doSBM(c, p, blkdata)

		c.SendCmd(&mt.ToCltBlkData{
			Blkpos: p,
			Blk:    blkdata.MapBlk,
		})
	}()
	return
}

// func GetBlk(p [3]int16) *mtmap.MapBlk)
// func SetBlk(p [3]int16, blk *mtmap.MapBlk)
// are in db.go

// GetNode returns the given mt.Content at a specified spot
// returns nil if blk does not exist
func GetNode(pos [3]int16) *mt.Node {
	p, i := mt.Pos2Blkpos(pos)
	blk := <-GetBlk(p)

	if blk == nil {
		return nil
	}

	return &mt.Node{
		Param0: blk.Param0[i],
		Param1: blk.Param1[i],
		Param2: blk.Param2[i],
	}
}

// SetNode reads a blk and sets node to node
// then saves
func SetNode(pos [3]int16, node mt.Node) {
	blk, i := mt.Pos2Blkpos(pos)
	oldBlk := <-GetBlk(blk)

	if oldBlk == nil {
		oldBlk = EmptyBlk()

		oldBlk.Flags |= mtmap.NotGenerated
	}

	oldBlk.Param0[i] = node.Param0
	oldBlk.Param1[i] = node.Param1
	oldBlk.Param2[i] = node.Param2
	// TODO: node meta

	SetBlk(blk, oldBlk)
}

// GetConfigFields returns a list of configuration fields incl a description
// TODO: write configuration lib
/*func GetConfigFileds() []struct{ Name, Desc string } {
	return []struct{ Name, Desc string }{
		struct{ Name, Desc string }{
			Name: "Load",
			Desc: "bool; if false disables spiral loading alogorithm; default true",
		},
	}
}*/
