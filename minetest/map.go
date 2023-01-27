package minetest

import (
	"github.com/anon55555/mt"

	"log"
	"strings"
	"sync"
	"time"
)

func init() {
	RegisterStage2(func() {
		data := GetConfig("map-driver")
		driver, ok := data.(string)
		if !ok {
			log.Fatalf("Error: Map: no driver specified config field 'map-driver' is empty!\nAvailable: %s\n", strings.Join(ListDrivers(), ", "))
		}

		driversMu.RLock()
		defer driversMu.RUnlock()

		activeDriver, ok = drivers[driver]
		if !ok {
			log.Fatalf("Error: Map: driver specified is invalid!\nAvailable: %s\n", strings.Join(ListDrivers(), ", "))
		}

		data = GetConfig("map-path")
		file, ok := data.(string)
		if !ok {
			log.Fatalf("Error: Map: no map specified config field 'map-path' is empty!\n")
		}

		err := activeDriver.Open(file)
		if err != nil {
			log.Fatalf("Error: %s\n", err)
		}

		MapLogger.Printf("Load DIM (0): %s(%s)\n", driver, file)
	})
}

func ListDrivers() (s []string) {
	driversMu.RLock()
	defer driversMu.RUnlock()

	for k := range drivers {
		s = append(s, k)
	}

	return
}

var (
	activeDriver MapDriver

	drivers   = make(map[string]MapDriver)
	driversMu sync.RWMutex
)

func RegisterMapDriver(name string, driver MapDriver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	drivers[name] = driver
}

type MapDriver interface {
	Open(string) error

	GetBlk([3]int16) (*MapBlk, error)
	SetBlk(*MapBlk) error
}

type MapBlk struct {
	MapBlk mt.MapBlk
	Pos    [3]int16

	Driver MapDriver

	sync.RWMutex

	ForceLoaded bool      // if set the default func for cleanup block won't be unloaded
	Loaded      time.Time // timestamp when blk was loaded (unixmillis)
	LastAccess  time.Time // timestamp when blk was last Accessed (unixmillis)
	LastSeen    time.Time // timestamp when client was in blk for the last time (unixmillis)
	LastRefresh time.Time // timestamp when blk was last manualy refreshed (unixmillis)

	deleted bool

	loadedByMu sync.RWMutex
	loadedBy   map[*Client]struct{}
}

func (blk *MapBlk) Save() error {
	return blk.Driver.SetBlk(blk)
}

func (blk *MapBlk) IsLoadedBy(c *Client) bool {
	blk.loadedByMu.RLock()
	defer blk.loadedByMu.RUnlock()

	_, ok := blk.loadedBy[c]
	return ok
}

func (blk *MapBlk) expired() bool {
	if expiredFuncs != nil {
		for _, f := range expiredFuncs {
			if !f(blk) {
				return false
			}
		}
	} else {
		if blk.ForceLoaded {
			return false
		}

		if blk.LastAccess.Add(time.Minute).Sub(time.Now()) <= 0 {
			return false
		}
	}

	return true
}

var (
	loaded   map[*Client]map[[3]int16]*MapBlk
	loadedMu sync.RWMutex
)

func markLoaded(clt *Client, p [3]int16) {
	loadedMu.Lock()
	defer loadedMu.Unlock()

	if loaded == nil {
		loaded = make(map[*Client]map[[3]int16]*MapBlk)
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
		blk.loadedBy = make(map[*Client]struct{})
	}

	blk.loadedBy[clt] = struct{}{}
}

func isLoaded(clt *Client, p [3]int16) bool {
	loadedMu.RLock()
	defer loadedMu.RUnlock()

	if loaded == nil || loaded[clt] == nil {
		return false
	} else if _, ok := loaded[clt][p]; !ok {
		return false
	}

	return true
}

func sendCltBlk(clt *Client, p [3]int16) <-chan struct{} {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	blk := GetBlk(p)

	blk.RLock()
	defer blk.RUnlock()
	ack, _ := clt.SendCmd(&mt.ToCltBlkData{
		Blkpos: p,
		Blk:    blk.MapBlk,
	})

	return ack
}
