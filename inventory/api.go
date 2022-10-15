package inventory

import (
	//	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"errors"
	"sync"
)

var (
	ErrorDefaultPlayerInventoryRegisterd    = errors.New("Default Inventory was  already registerd.")
	ErrorDefaultPlayerInventoryNotRegisterd = errors.New("Default Player Inventory type was not registerd")
)

var (
	playerInventoryTypes   = make(map[string]func(*minetest.Client) PlayerInv)
	playerInventoryTypesMu sync.RWMutex
)

func RegisterPlayerInventoryType(name string, maker func(*minetest.Client) PlayerInv) {
	playerInventoryTypesMu.Lock()
	defer playerInventoryTypesMu.Unlock()

	playerInventoryTypes[name] = maker
}

var (
	defaultInventory   string
	defaultInventoryMu sync.RWMutex
)

func SetDefaultInventory(name string) error {
	defaultInventoryMu.Lock()
	defer defaultInventoryMu.Unlock()

	if defaultInventory != "" {
		return ErrorDefaultPlayerInventoryRegisterd
	}

	playerInventoryTypesMu.RLock()
	defer playerInventoryTypesMu.RUnlock()
	if _, ok := playerInventoryTypes[name]; !ok {
		return ErrorDefaultPlayerInventoryNotRegisterd
	}

	defaultInventory = name

	return nil
}
