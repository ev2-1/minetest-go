package minetest

import (
	"github.com/anon55555/mt"

	"bytes"
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
func LoadBlk(clt *Client, p [3]int16) <-chan struct{} {
	if isLoaded(clt, p) {
		return nil
	}

	TryCache(p)
	ack := sendCltBlk(clt, p)

	go func() {
		<-ack
		markLoaded(clt, p)

	}()

	return ack
}

// GetBlk returns a pointer to block at a BlkPos
func GetBlk(p [3]int16) *MapBlk {
	if ConfigVerbose() {
		MapLogger.Printf("GetBlk(%d,%d,%d)\n", p[0], p[1], p[2])
	}

	if err := TryCache(p); err != nil {
		panic(err)
	}

	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	return mapCache[p]
}

// GetNode returns a mt.Node and NodeMeta for a coordinate
// If no NodeMeta is specified returns mt.Node and nil
func GetNode(p [3]int16) (node mt.Node, meta *mt.NodeMeta) {
	blk, i := mt.Pos2Blkpos(p)

	mapblk := GetBlk(blk).MapBlk

	return mt.Node{
		Param0: mapblk.Param0[i],
		Param1: mapblk.Param1[i],
		Param2: mapblk.Param2[i],
	}, mapblk.NodeMetas[i]
}

// SetNode sets a mt.Node and NodeMeta for a coordinate
// If no NodeMeta is specified it WILL be overwritten
func SetNode(p [3]int16, node mt.Node, meta *mt.NodeMeta) {
	MapLogger.Printf("SetNode (%d,%d,%d) mt.Content(%d)", p[0], p[1], p[2], node.Param0)

	blk, i := mt.Pos2Blkpos(p)

	mapblk := GetBlk(blk)
	mapblk.Lock()
	defer mapblk.Unlock()

	mtblk := mapblk.MapBlk

	update := false
	// check if anything will update
	if mtblk.Param0[i] != node.Param0 || mtblk.Param1[i] != node.Param1 || mtblk.Param2[i] != node.Param2 {
		update = true
	}

	keepMeta := NodeMetasEqual(meta, mtblk.NodeMetas[i])

	if update {
		BroadcastClientM(mapblk.loadedBy, &mt.ToCltAddNode{
			Pos:      p,
			Node:     node,
			KeepMeta: keepMeta,
		})
	}

	if !keepMeta {
		BroadcastClientM(mapblk.loadedBy, &mt.ToCltNodeMetasChanged{
			Changed: map[[3]int16]*mt.NodeMeta{
				p: meta,
			},
		})
	}

	mtblk.Param0[i] = node.Param0
	mtblk.Param1[i] = node.Param1
	mtblk.Param2[i] = node.Param2

	if meta == nil {
		delete(mtblk.NodeMetas, i)
	} else {
		mtblk.NodeMetas[i] = meta
	}

	mapblk.MapBlk = mtblk

	return
}

func Fields2map(s []mt.NodeMetaField) map[string]mt.NodeMetaField {
	m := map[string]mt.NodeMetaField{}

	for _, field := range s {
		m[field.Name] = field
	}

	return m
}

func NodeMetasEqual(m1, m2 *mt.NodeMeta) bool {
	if m1 == nil && m2 == nil {
		return true
	}

	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	m1.Inv.Serialize(buf1)
	m2.Inv.Serialize(buf2)

	if buf1.String() != buf2.String() {
		return false
	}

	// compare Fields
	fields1 := Fields2map(m1.Fields)
	fields2 := Fields2map(m2.Fields)

	if len(fields1) != len(fields2) {
		return false
	}

	for k, v := range fields1 {
		if fields2[k] != v {
			return false
		}
	}

	return true
}
