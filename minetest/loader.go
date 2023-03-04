package minetest

import (
	"sync"
)

type MapLoader interface {
	Make(*Client) MapLoader

	Load()
}

var (
	mapLoadersMu sync.RWMutex
	mapLoaders   = map[string]*Registerd[MapLoader]{}

	defaultMapLoader   string = GetConfigV("default-map-loader", "debug")
	defaultMapLoaderMu sync.RWMutex
)

func RegisterMapLoader(name string, loader MapLoader) {
	mapLoadersMu.Lock()
	defer mapLoadersMu.Unlock()

	mapLoaders[name] = &Registerd[MapLoader]{Caller(1), loader}
}

func GetMapLoader(name string) *Registerd[MapLoader] {
	mapLoadersMu.RLock()
	defer mapLoadersMu.RUnlock()

	return mapLoaders[name]
}

func DefaultMapLoader() string {
	defaultMapLoaderMu.RLock()
	defer defaultMapLoaderMu.RUnlock()

	return defaultMapLoader
}
