package mmap

import (
	"fmt"
	"log"
	"sync"

	"time"
)

var mapCache map[[3]int16]*MapBlk
var mapCacheMu sync.RWMutex

// IsCached returns true if there is a valid cache for pos p
func IsCached(pos [3]int16) bool {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	_, f := mapCache[pos]

	return f
}

// TryCache caches if mapblk is either not cached already
// if cache is still valid, does nothing
// Refreshes Loaded.
func TryCache(pos [3]int16) error {
	if !IsCached(pos) {
		loadIntoCache(pos)
	}

	return nil
}

func loadIntoCache(pos [3]int16) error {
	mapCacheMu.Lock()
	defer mapCacheMu.Unlock()

	blk, err := activeDriver.GetBlk(pos)
	if err != nil {
		return err
	}

	if mapCache == nil {
		mapCache = make(map[[3]int16]*MapBlk)
	}

	mapCache[pos] = MakeMapBlk(blk, pos)

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

	for i := 0; i < len(delQueue); i++ {
		delete(mapCache, delQueue[i])
	}
}

// enumerateExpiredBlks enumerates all blks that should be unloaded
func enumerateExpiredBlks() (s [][3]int16) {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	for pos, blk := range mapCache {
		blk.Lock()

		// check if pos matches:
		if pos != blk.MapBlk.Pos() {
			log.Fatal(fmt.Sprintf("mapblk dosn't have correct pos has %v, was expecting %v", blk.MapBlk.Pos(), pos))
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
