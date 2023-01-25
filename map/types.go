package mmap

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"sync"
	"time"
)

var (
	activeDriver MapDriver

	drivers   = make(map[string]MapDriver)
	driversMu sync.RWMutex
)

func RegisterDriver(name string, driver MapDriver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	drivers[name] = driver
}

type MapDriver interface {
	Open(string) error

	GetBlk([3]int16) (DriverMapBlk, error)
	SetBlk([3]int16, *MapBlk) error
}

type DriverMapBlk interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()

	MapBlk() *mt.MapBlk
	Pos() [3]int16

	Save(*MapBlk) error
}

type MapBlk struct {
	MapBlk DriverMapBlk
	sync.RWMutex

	ForceLoaded bool      // if set the default func for cleanup block won't be unloaded
	Loaded      time.Time // timestamp when blk was loaded (unixmillis)
	LastAccess  time.Time // timestamp when blk was last Accessed (unixmillis)
	LastSeen    time.Time // timestamp when client was in blk for the last time (unixmillis)
	LastRefresh time.Time // timestamp when blk was last manualy refreshed (unixmillis)

	deleted bool

	loadedByMu sync.RWMutex
	loadedBy   map[*minetest.Client]struct{}
}

func MakeMapBlk(blk DriverMapBlk, pos [3]int16) *MapBlk {
	now := time.Now()

	return &MapBlk{
		MapBlk: blk,

		Loaded:      now,
		LastRefresh: now,
		LastAccess:  now,

		loadedBy: make(map[*minetest.Client]struct{}),
	}
}

func (blk *MapBlk) IsLoadedBy(c *minetest.Client) bool {
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
