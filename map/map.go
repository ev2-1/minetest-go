package mmap

import (
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"strings"
)

func init() {
	minetest.RegisterStage2(func() {
		data := minetest.GetConfig("map-driver")
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

		data = minetest.GetConfig("map-path")
		file, ok := data.(string)
		if !ok {
			log.Fatalf("Error: Map: no map specified config field 'map-path' is empty!\n")
		}

		err := activeDriver.Open(file)
		if err != nil {
			log.Fatalf("Error: %s\n", err)
		}

		log.Printf("Map: %s(%s)\n", driver, file)
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
