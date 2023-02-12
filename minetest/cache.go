package minetest

import (
	"fmt"
	"log"
	"sync"

	"time"
)

var mapCache map[IntPos]*MapBlk
var mapCacheMu sync.RWMutex

// IsCached returns true if there is a valid cache for pos p
func IsCached(pos IntPos) bool {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	_, f := mapCache[pos]

	return f
}

// TryCache caches if mapblk is either not cached already
// if cache is still valid, does nothing
// Refreshes Loaded.
func TryCache(pos IntPos) error {
	if !IsCached(pos) {
		return loadIntoCache(pos)
	}

	return nil
}

func loadIntoCache(pos IntPos) error {
	if ConfigVerbose() {
		MapLogger.Printf("Loading (%d,%d,%d) %s (%d) into cache\n", pos.Pos[0], pos.Pos[1], pos.Pos[2], pos.Dim, pos.Dim)
	}

	mapCacheMu.Lock()
	defer mapCacheMu.Unlock()

	blk, err := mapIO[pos.Dim].GetBlk(pos.Pos)
	if err != nil {
		return err
	}

	if mapCache == nil {
		mapCache = make(map[IntPos]*MapBlk)
	}

	mapCache[pos] = blk

	return nil
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
