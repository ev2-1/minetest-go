package minetest

import (
	"github.com/anon55555/mt"

	"bytes"
	"errors"
	"fmt"
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
func LoadBlk(clt *Client, p IntPos) <-chan struct{} {
	mapCacheMu.Lock()
	defer mapCacheMu.Unlock()

	if isLoaded(clt, p) {
		return nil
	}

	err := tryCache(p)
	if err != nil {
		clt.Logf("WARN: TryCache: %s\n", err)
		if errors.Is(ErrInvalidDim, err) {
			clt.Logf("WARN: %s: resetting dimension to DIM0!\n", err)
			pos := clt.GetPos()
			pos.Dim = 0
			SetPos(clt, pos, false)
		}
	}

	ack := sendCltBlk(clt, p)

	go func() {
		<-ack

		markLoaded(clt, p)
	}()

	return ack
}

// GetBlk returns a pointer to block at a BlkPos
func GetBlk(p IntPos) *MapBlk {
	mapCacheMu.Lock()
	defer mapCacheMu.Unlock()

	return getBlk(p)
}

func getBlk(p IntPos) *MapBlk {
	if ConfigVerbose() {
		Loggers.Verbosef("GetBlk(%s)\n", 1, p)
	}

	if err := tryCache(p); err != nil {
		panic(err)
	}

	return mapCache[p]

}

func Pos2Blkpos(p IntPos) (ni IntPos, i uint16) {
	ni.Pos, i = mt.Pos2Blkpos(p.Pos)
	ni.Dim = p.Dim

	return ni, i
}

func Blkpos2Pos(p IntPos, i uint16) (ni IntPos) {
	return IntPos{
		Pos: mt.Blkpos2Pos(p.Pos, i),
		Dim: p.Dim,
	}
}

// GetNode returns a mt.Node and NodeMeta for a coordinate
// If no NodeMeta is specified returns mt.Node and nil
func GetNode(p IntPos) (node mt.Node, meta *mt.NodeMeta) {
	blk, i := Pos2Blkpos(p)

	mapblk := GetBlk(blk).MapBlk

	return mt.Node{
		Param0: mapblk.Param0[i],
		Param1: mapblk.Param1[i],
		Param2: mapblk.Param2[i],
	}, mapblk.NodeMetas[i]
}

// SetNode sets a mt.Node and NodeMeta for a coordinate
// If no NodeMeta is specified it WILL be overwritten
func SetNode(p IntPos, node mt.Node, meta *mt.NodeMeta) {
	Loggers.Verbosef("SetNode (%s) mt.Content(%d)", 1, p, node.Param0)

	blk, i := Pos2Blkpos(p)

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
			Pos:      p.Pos,
			Node:     node,
			KeepMeta: keepMeta,
		})
	}

	if !keepMeta && meta != nil {

		BroadcastClientM(mapblk.loadedBy, &mt.ToCltNodeMetasChanged{
			Changed: map[[3]int16]*mt.NodeMeta{
				p.Pos: meta,
			},
		})
	}

	mtblk.Param0[i] = node.Param0
	mtblk.Param1[i] = node.Param1
	mtblk.Param2[i] = node.Param2

	if meta == nil {
		delete(mtblk.NodeMetas, i)
	} else {
		if mtblk.NodeMetas == nil {
			mtblk.NodeMetas = make(map[uint16]*mt.NodeMeta)
		}
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
	if m1 == nil {
		if m2 == nil {
			return true
		} else {
			return false
		}
	}

	if m2 == nil {
		return false
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

func Map2Slice[V comparable](m map[V]struct{}) []V {
	s := make([]V, len(m))

	var i int

	for k := range m {
		s[i] = k

		i++
	}

	return s
}

func strSlice[V any](s []V) []string {
	strs := make([]string, len(s))
	for k := range strs {
		strs[k] = fmt.Sprintf("%v", s[k])
	}

	return strs
}
