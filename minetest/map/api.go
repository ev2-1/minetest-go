package mmap

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"sync"
)

var expiredFuncs []func(*MapBlk) bool
var expiredFuncMu sync.RWMutex

// add a function that is used to determin whether a blk can be unloaded
// all have to be false, to unload a blk
// its possible for a chunk to get unloaded without your function getting called, for a hook use `RegisterUnloadHook`
func AddExpiredCondition(f func(*MapBlk) bool) {
	expiredFuncMu.Lock()
	defer expiredFuncMu.Unlock()

	expiredFuncs = append(expiredFuncs, f)
}

// LoadBlk sends a blk and marks it as send
// only sends updates after that until client send DeletedBlks
func LoadBlk(clt *minetest.Client, p [3]int16) <-chan struct{} {
	if isLoaded(clt, p) {
		return nil
	}

	ch := make(chan struct{})

	go func() {
		loadIntoCache(p)
		<-sendCltBlk(clt, p)
		markLoaded(clt, p)

		close(ch)
	}()

	return ch
}

// GetBlk returns a pointer to block at a BlkPos
func GetBlk(p [3]int16) *MapBlk {
	TryCache(p)

	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	return mapCache[p]
}

// GetNode returns a mt.Node and NodeMeta for a coordinate
// If no NodeMeta is specified returns mt.Node and nil
func GetNode(p [3]int16) (node mt.Node, meta *mt.NodeMeta) {
	blk, i := mt.Pos2Blkpos(p)

	mapblk := GetBlk(blk).MapBlk.MapBlk()

	return mt.Node{
		Param0: mapblk.Param0[i],
		Param1: mapblk.Param1[i],
		Param2: mapblk.Param2[i],
	}, mapblk.NodeMetas[i]
}

// SetNode sets a mt.Node and NodeMeta for a coordinate
// If no NodeMeta is specified it WILL be overwritten
func SetNode(p [3]int16, node mt.Node, meta *mt.NodeMeta) {
	blk, i := mt.Pos2Blkpos(p)

	mapblk := GetBlk(blk).MapBlk.MapBlk()
	mapblk.Param0[i] = node.Param0
	mapblk.Param1[i] = node.Param1
	mapblk.Param2[i] = node.Param2

	if meta == nil {
		delete(mapblk.NodeMetas, i)
	} else {
		mapblk.NodeMetas[i] = meta
	}

	return
}
