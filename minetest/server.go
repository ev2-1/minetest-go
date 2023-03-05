/*
minetest-go is a *very* simple minetest server written in golang
*/
package minetest

import (
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

const (
	SerializeVer     = 28
	ProtoVer         = 39
	VersionString    = "mtgo-5.4.1"
	MaxPlayerNameLen = 20
)

var playerNameChars = regexp.MustCompile("^[a-zA-Z0-9-_]+$")

var execDir string
var execDirOnce sync.Once

// gets path relative to executable
func Path(path string) string {
	execDirOnce.Do(func() {
		f, err := os.Executable()
		if err != nil {
			Loggers.Errorf("%s", 1, err)
			os.Exit(1)
		}

		execDir = filepath.Dir(f)
	})

	return execDir + "/" + path
}

type ServerState uint8

var stateMu sync.RWMutex
var state ServerState = StateInitializing

func State() ServerState {
	stateMu.RLock()
	defer stateMu.RUnlock()

	return state
}

const (
	StateInitializing ServerState = iota
	StateOnline
	StateShuttingDown
	StateOffline // only reached when saving
)
