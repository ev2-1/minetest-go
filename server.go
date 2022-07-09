/*
	minetest-go is a *very* simple minetest server written in golang
*/
package minetest

import (
	"log"
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
			log.Fatal(err)
		}

		execDir = filepath.Dir(f)
	})

	return execDir + "/" + path
}
