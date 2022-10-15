package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"encoding/json"
	"log"
	"os"
	"sync"
)

var (
	storage   map[string]*mt.Inv
	storageMu sync.RWMutex
)

func init() {
	storage = make(map[string]*mt.Inv)
	f, err := os.OpenFile(minetest.Path("player_inv.json"), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal("error opening 'player_inv.json'", err)
	}

	defer f.Close()

	d := json.NewDecoder(f)

	storageMu.Lock()
	defer storageMu.Unlock()
	d.Decode(&storage)
}

func init() {
	minetest.RegisterShutdownHooks(func() {
		log.Println("Writing player inventorys")

		storageMu.Lock()
		f, err := os.OpenFile(minetest.Path("player_inv.json"), os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			log.Println("error opening file 'player_inv.json'", err)
			return
		}

		defer f.Close()

		e := json.NewEncoder(f)

		e.Encode(storage)
	})
}
