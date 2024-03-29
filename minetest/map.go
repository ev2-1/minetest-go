package minetest

import (
	"github.com/anon55555/mt"

	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

func init() {
	RegisterStage2(func() {
		drv, ok := GetConfig("map-driver", "")
		if !ok {
			Loggers.Errorf("Error: Map: no driver specified config field 'map-driver' is empty!\nAvailable: %s\n", 1, strings.Join(ListDrivers(), ", "))
			os.Exit(1)
		}

		driversMu.RLock()
		driver, ok := drivers[drv]
		if !ok {
			Loggers.Errorf("Error: Map: driver specified is invalid!\nAvailable: %s\n", 1, strings.Join(ListDrivers(), ", "))
			os.Exit(1)
		}

		driversMu.RUnlock()

		driver = driver.Make()

		file, ok := GetConfig("map-path", "map.sqlite")
		if !ok {
			Loggers.Errorf("Error: Map: no map specified config field 'map-path' is empty!\n", 1)
			os.Exit(1)
		}

		err := driver.Open(file)
		if err != nil {
			Loggers.Errorf("Error: Can't open '%s' for DIM0: %s", 1, file, err)
		}

		mapgen, ok := GetConfig("map-generator", "flat")
		if !ok {
			Loggers.Errorf("Error: Map: no map specified config field 'map-path' is empty!\n", 1)
			os.Exit(1)
		}

		mapargs, ok := GetConfig("map-generator-args", "")
		if !ok {
			Loggers.Errorf("Error: Map: no map specified config field 'map-path' is empty!\n", 1)
		}

		generatorsMu.RLock()
		generator, ok := generators[mapgen]
		if !ok {
			Loggers.Errorf("Error for DIM0: %s\n", 1, ErrInvalidGenerator)
			os.Exit(1)
		}

		generatorsMu.RUnlock()
		generator = generator.Make(driver, mapargs)

		//open default DIM0:
		err = registerDim(&Dimension{
			Driver:    driver,
			Generator: generator,
			Name:      "DIM0",
			ID:        0,
		})

		if err != nil {
			Loggers.Errorf("Error: Map: can't create DIM0: %s\n", 1, err)
			os.Exit(1)
		}
	})

	RegisterPktProcessor(func(c *Client, pkt *mt.Pkt) {
		if blks, ok := pkt.Cmd.(*mt.ToSrvDeletedBlks); ok {
			for _, blk := range blks.Blks {
				p := c.GetPos().IntPos()
				p.Pos = blk

				markUnloaded(c, p)
			}
		}
	})
}

var ErrInvalidDriver = errors.New("invalid mapdriver")
var ErrInvalidGenerator = errors.New("invalid map generator")

// NewDim creates a new Dimension and registers it
func NewDim(name, gen, genargs, drv, file string) (*Dimension, error) {
	driversMu.RLock()
	driver, ok := drivers[drv]
	if !ok {
		return nil, ErrInvalidDriver
	}

	driversMu.RUnlock()

	generatorsMu.RLock()
	generator, ok := generators[gen]
	if !ok {
		return nil, ErrInvalidGenerator
	}

	generatorsMu.RUnlock()

	//	gen, ok = drivers[drv]
	//	if !ok {
	//		return nil, ErrInvalidDriver
	//	}

	dim := &Dimension{
		Driver: driver.Make(),
		Name:   name,
		ID:     0, // Aquire later automatically
	}

	dim.Generator = generator.Make(dim.Driver, genargs)

	err := dim.Driver.Open(file)
	if err != nil {
		return nil, err
	}

	_, err = RegisterDim(dim)
	if err != nil {
		return nil, err
	}

	return dim, err
}

// RegisterDim registers a new Dimension
// Id will be generated by NewDim if id == 0; returns id
func RegisterDim(d *Dimension) (DimID, error) {
	if d.ID == 0 {
		dimensionsMu.Lock()
		d.ID = dimCounter
		dimCounter++
		dimensionsMu.Unlock()
	}

	err := registerDim(d)
	if err != nil {
		Loggers.Defaultf("WARN: failed to register new dimension %s (%d)\n", 1, d.Name, d.ID)

		return 0, nil
	}

	return d.ID, nil
}

type Dimension struct {
	sync.RWMutex

	Driver    MapDriver
	Generator MapGenerator
	Name      string
	ID        DimID
}

func registerDim(d *Dimension) error {
	Loggers.Defaultf("Registering DIM %s (%d): though %T\n", 1, d.Name, d.ID, d.Driver)

	dimensionsMu.Lock()
	dimensions[d.Name] = d
	dimensionsR[d.ID] = d
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
	dimensions         = make(map[string]*Dimension)
	dimensionsR        = make(map[DimID]*Dimension)
	dimCounter   DimID = 1

	drivers   = make(map[string]MapDriver)
	driversMu sync.RWMutex

	generators   = make(map[string]MapGenerator)
	generatorsMu sync.RWMutex
)

// RegisterMapGenerator `gen` for `name`
// overwrites if allready exists
func RegisterMapGenerator(name string, gen MapGenerator) {
	generatorsMu.Lock()
	defer generatorsMu.Unlock()

	generators[name] = gen
}

// MapGenerator specifies the API a MapGen has to use
// has access to mapdriver
// can but should (in most cases) not set blocks its not asked to
type MapGenerator interface {
	Make(drv MapDriver, args string) MapGenerator

	// Generate is called with a BlkPos
	// Should save generated Blk into MapDriver
	Generate([3]int16) (*MapBlk, error)
}

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
		Loggers.Defaultf("[%s] marking [%v] loaded\n", 1, clt, p)
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
	mapCacheMu.RUnlock()

	if blk == nil {
		Loggers.Default("blk is nil while marking loaded\n", 1)
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
	dim := clt.GetFullPos().OldPos.Dim
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

	return Acks(acks...), nil
}

func isLoaded(clt *Client, p IntPos) bool {
	pos := clt.GetPos()

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
	blk := getBlk(p)

	blk.RLock()
	defer blk.RUnlock()
	ack, _ := clt.SendCmd(&mt.ToCltBlkData{
		Blkpos: p.Pos,
		Blk:    blk.MapBlk,
	})

	return ack
}
