package minetest

import (
	"github.com/anon55555/mt"

	"errors"
	"log"
	"strings"
	"sync"
	"time"
)

func init() {
	RegisterStage2(func() {
		driver, ok := GetConfigString("map-driver", "")
		if !ok {
			log.Fatalf("Error: Map: no driver specified config field 'map-driver' is empty!\nAvailable: %s\n", strings.Join(ListDrivers(), ", "))
		}

		driversMu.RLock()
		defer driversMu.RUnlock()

		defaultDriver, ok = drivers[driver]
		if !ok {
			log.Fatalf("Error: Map: driver specified is invalid!\nAvailable: %s\n", strings.Join(ListDrivers(), ", "))
		}

		file, ok := GetConfigString("map-path", "map.sqlite")
		if !ok {
			log.Fatalf("Error: Map: no map specified config field 'map-path' is empty!\n")
		}

		//open default DIM0:
		err := newDim(0, "DIM0", defaultDriver, file)
		if err != nil {
			log.Fatalf("Error: Map: can't create DIM0: %s\n", err)
		}
	})
}

var ErrInvalidDriver = errors.New("invalid mapdriver")

func NewDim(name, drv, file string) (Dim, error) {
	driversMu.RLock()

	driver, ok := drivers[drv]
	if !ok {
		return 0, ErrInvalidDriver
	}

	driversMu.RUnlock()

	dimensionsMu.Lock()
	id := dimCounter
	dimCounter++
	dimensionsMu.Unlock()

	err := newDim(id, name, driver, file)
	if err != nil {
		MapLogger.Printf("WARN: failed to register new dimension %s (%d)\n", name, id)

		return 0, nil
	}

	return id, nil
}

func newDim(d Dim, name string, drv MapDriver, file string) error {
	driver := drv.Make()

	err := driver.Open(file)
	if err != nil {
		return err
	}

	MapLogger.Printf("Load DIM %s (%d): though %s@%T\n", name, d, file, drv)
	mapIOMu.Lock()
	mapIO[d] = driver
	mapIOMu.Unlock()

	dimensionsMu.Lock()
	dimensions[name] = d
	dimensionsR[d] = name
	dimensionsMu.Unlock()

	return nil
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
	defaultDriver MapDriver

	dimensionsMu sync.RWMutex
	dimensions       = make(map[string]Dim)
	dimensionsR      = make(map[Dim]string)
	dimCounter   Dim = 1

	mapIO   = make(map[Dim]MapDriver)
	mapIOMu sync.RWMutex

	drivers   = make(map[string]MapDriver)
	driversMu sync.RWMutex
)

func RegisterMapDriver(name string, driver MapDriver) {
	driversMu.Lock()
	defer driversMu.Unlock()

	drivers[name] = driver
}

type MapDriver interface {
	Make() MapDriver // create new instance of MapDriver

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

func markLoaded(clt *Client, p IntPos) {
	loadedMu.Lock()
	defer loadedMu.Unlock()

	if ConfigVerbose() {
		clt.Logf("marking [%v] loaded\n", p)
	}

	if loaded == nil {
		loaded = make(map[*Client]map[[3]int16]*MapBlk)
	}

	if loaded[clt] == nil {
		loaded[clt] = make(map[[3]int16]*MapBlk)
	}

	mapCacheMu.RLock()
	blk := mapCache[p]
	loaded[clt][p.Pos] = blk
	defer mapCacheMu.RUnlock()

	if blk == nil {
		MapLogger.Println("blk is nil while marking loaded")
		return
	}

	blk.Lock()
	defer blk.Unlock()

	if blk.loadedBy == nil {
		blk.loadedBy = make(map[*Client]struct{})
	}

	blk.loadedBy[clt] = struct{}{}
}

func markUnloaded(clt *Client, p IntPos) {
	loadedMu.Lock()
	defer loadedMu.Unlock()

	if ConfigVerbose() {
		clt.Logf("marking [%v] unloaded\n", p)
	}

	if loaded == nil {
		loaded = make(map[*Client]map[[3]int16]*MapBlk)
	}

	if loaded[clt] == nil {
		loaded[clt] = make(map[[3]int16]*MapBlk)
	}

	blk := GetBlk(p)
	blk.Lock()
	defer blk.Unlock()

	if blk.loadedBy == nil {
		blk.loadedBy = make(map[*Client]struct{})
	}

	delete(blk.loadedBy, clt)

	delete(loaded[clt], p.Pos)
}

var ignoreblk mt.MapBlk = func() (blk mt.MapBlk) {
	for k := range blk.Param0 {
		blk.Param0[k] = mt.Ignore
	}

	return
}()

func enumerateLoaded(c *Client) [][3]int16 {
	loadedMu.RLock()
	defer loadedMu.RUnlock()

	var s = make([][3]int16, len(loaded[c]))
	var i int

	for pos := range loaded[c] {
		s[i] = pos

		i++
	}

	return s
}

func unloadAll(clt *Client) (_ <-chan struct{}, err error) {
	dim := GetFullPos(clt).OldPos.Dim
	acks := make([]<-chan struct{}, len(loaded[clt]))
	var i int

	// Send client dummy blks
	for _, pos := range enumerateLoaded(clt) {
		clt.Logf("Overwriting blk(%v)\n", pos)
		aack, err := clt.SendCmd(&mt.ToCltBlkData{
			Blkpos: pos,
			Blk:    ignoreblk,
		})

		if err != nil {
			return nil, err
		}

		acks[i] = aack

		go func(pos [3]int16) {
			<-aack // wait for client to ack

			markUnloaded(clt, IntPos{pos, dim})
		}(pos)

		i++
	}

	aaack := make(chan struct{})
	go Acks(aaack, acks...)

	return aaack, nil
}

func isLoaded(clt *Client, p IntPos) bool {
	loadedMu.RLock()
	defer loadedMu.RUnlock()

	pos := GetPos(clt)

	if p.Dim != pos.Dim {
		return false
	}

	if loaded == nil || loaded[clt] == nil {
		return false
	} else if _, ok := loaded[clt][p.Pos]; !ok {
		return false
	}

	return true
}

func sendCltBlk(clt *Client, p IntPos) <-chan struct{} {
	mapCacheMu.RLock()
	defer mapCacheMu.RUnlock()

	blk := GetBlk(p)

	blk.RLock()
	defer blk.RUnlock()
	ack, _ := clt.SendCmd(&mt.ToCltBlkData{
		Blkpos: p.Pos,
		Blk:    blk.MapBlk,
	})

	return ack
}
