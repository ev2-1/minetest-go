package main

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"
	"io"
	"log"
	"os"

	"encoding/json"
)

func init() {
	// create folder (if not exists)
	path := minetest.Path("itemdefs/")
	os.Mkdir(path, 0777)

	dir, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range dir {
		f, err := os.Open(minetest.Path("itemdefs/" + file.Name()))
		if err != nil {
			log.Fatal(err)
		}
		parseFile(f)
	}
}

func parseFile(r io.Reader) {
	d := json.NewDecoder(r)

	var defs []*mt.NodeDef

	err := d.Decode(&defs)
	if err != nil {
		log.Printf("Error parsing nodedef '%s'", err.Error())
	}

	minetest.AddNodeDef(defs...)
}
