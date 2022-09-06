package mmap

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"sync"
)

var (
	loaded   map[*minetest.Client]map[[3]int16]*MapBlk
	loadedMu sync.RWMutex
)

func markLoaded(clt *minetest.Client, p [3]int16) {
	loadedMu.Lock()
	defer loadedMu.Unlock()

	if loaded == nil {
		loaded = make(map[*minetest.Client]map[[3]int16]*MapBlk)
	}

	if loaded[clt] == nil {
		loaded[clt] = make(map[[3]int16]*MapBlk)
	}

	mapCacheMu.RLock()
	blk := mapCache[p]
	loaded[clt][p] = blk
	defer mapCacheMu.RUnlock()

	if blk == nil {
		log.Println("[map] blk is nil while marking loaded")
		return
	}

	blk.Lock()
	defer blk.Unlock()

	if blk.loadedBy == nil {
		blk.loadedBy = make(map[*minetest.Client]struct{})
	}

	blk.loadedBy[clt] = struct{}{}
}

func isLoaded(clt *minetest.Client, p [3]int16) bool {
	loadedMu.RLock()
	defer loadedMu.RUnlock()

	if loaded == nil || loaded[clt] == nil {
		return false
	} else if _, ok := loaded[clt][p]; !ok {
		return false
	}

	return true
}

func sendCltBlk(clt *minetest.Client, p [3]int16) <-chan struct{} {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	blk, ok := mapCache[p]
	if !ok {
		return ch()
	}

	blk.RLock()
	defer blk.RUnlock()
	ack, _ := clt.SendCmd(&mt.ToCltBlkData{
		Blkpos: p,
		Blk:    blk.MapBlk.MapBlk,
	})

	return ack
}
