package minetest

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"time"
)

var mapCache map[IntPos]*MapBlk
var mapCacheMu sync.RWMutex

var ErrInvalidDim = errors.New("invalid dimension")

// IsCached returns true if there is a valid cache for pos p
func IsCached(pos IntPos) bool {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	return isCached(pos)
}

func isCached(pos IntPos) bool {
	_, f := mapCache[pos]

	return f
}

// TryCache caches if mapblk is either not cached already
// if cache is still valid, does nothing
// Refreshes Loaded.
func TryCache(pos IntPos) error {
	mapCacheMu.Lock()
	defer mapCacheMu.Unlock()

	return tryCache(pos)
}

func tryCache(pos IntPos) error {
	if !isCached(pos) {
		return loadIntoCache(pos)
	}

	return nil
}

// loads Blks at pos into Cache
// generates new if no Blk exists
func loadIntoCache(pos IntPos) error {
	if ConfigVerbose() {
		MapLogger.Printf("Loading (%d,%d,%d) %s (%d) into cache\n", pos.Pos[0], pos.Pos[1], pos.Pos[2], pos.Dim, pos.Dim)
	}

	dim := pos.Dim.Lookup()
	if dim == nil {
		MapLogger.Printf("Tired to access dimension %d, but is not registerd!\n", pos.Dim)
		return ErrInvalidDim
	}

	drv := dim.Driver

	blk, err := drv.GetBlk(pos.Pos)
	if err != nil {
		MapLogger.Printf("Info: error encounterd in GetBlk: [%v]: %s\n", pos, err)
	}

	if mapCache == nil {
		mapCache = make(map[IntPos]*MapBlk)
	}

	if blk == nil {
		MapLogger.Printf("blk at [%v] does not exists. generating\n", pos)
		blk, err = dim.Generator.Generate(pos.Pos)
		if err != nil {
			return err
		}
	}

	mapCache[pos] = blk
	go mapALH(pos, blk)

	return nil
}

// map after load hooks
type ALH func(IntPos, *MapBlk)

// mapALH
var (
	mapALHs   = make(map[*Registerd[ALH]]struct{})
	mapALHsMu sync.RWMutex
)

func RegisterALH(h ALH) HookRef[Registerd[ALH]] {
	mapALHsMu.Lock()
	defer mapALHsMu.Unlock()

	r := &Registerd[ALH]{Caller(1), h}
	ref := HookRef[Registerd[ALH]]{&mapALHsMu, mapALHs, r}

	mapALHs[r] = struct{}{}

	return ref
}

func mapALH(pos IntPos, blk *MapBlk) {
	mapALHsMu.RLock()
	defer mapALHsMu.RUnlock()

	for alh := range mapALHs {
		alh.Thing(pos, blk)
	}
}

// CleanCache cleans the cache of expired blks
func CleanCache() {
	delQueue := enumerateExpiredBlks()
	if delQueue == nil {
		return
	}

	mapCacheMu.Lock()
	defer mapCacheMu.Unlock()

	var unloaded int

	for i := 0; i < len(delQueue); i++ {
		if ConfigVerbose() {
			p := delQueue[i]
			MapLogger.Printf("Unloading (%d,%d,%d) %s (%d)",
				p.Pos[0], p.Pos[1], p.Pos,
				p.Dim, p.Dim,
			)
		}

		blk, ok := mapCache[delQueue[i]]
		if ok {
			blk.Save()
		}
		delete(mapCache, delQueue[i])
	}

	if unloaded > 0 {
		MapLogger.Printf("Unloaded %d chunks", unloaded)
	}
}

func enumerateLoadedBlks() (s []IntPos) {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	for pos, _ := range mapCache {
		s = append(s, pos)
	}

	return
}

func SaveCache() {
	loaded := enumerateLoadedBlks()

	for _, pos := range loaded {
		GetBlk(pos).Save()
	}
}

func init() {
	RegisterSaveFileHook(func() {
		MapLogger.Printf("Saving to disk")
		SaveCache()
	})
}

// enumerateExpiredBlks enumerates all blks that should be unloaded
func enumerateExpiredBlks() (s []IntPos) {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	for pos, blk := range mapCache {
		blk.Lock()

		// check if pos matches:
		if pos.Pos != blk.Pos {
			log.Fatal(fmt.Sprintf("mapblk dosn't have correct pos has %v, was expecting %v", blk.Pos, pos))
		}

		if blk.expired() {
			s = append(s, pos)
		}

		blk.Unlock()
	}

	return
}

func init() {
	go func() {
		ticker := time.NewTicker(time.Second * 10)

		for range ticker.C {
			CleanCache()
		}
	}()
}
