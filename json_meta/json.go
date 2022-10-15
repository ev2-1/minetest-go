package json_nodemeta

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	"io"
	"log"
	"os"

	"encoding/json"
)

func init() {
	// create folder (if not exists)
	path := minetest.Path("nodedefs/")
	os.Mkdir(path, 0777)

	dir, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range dir {
		f, err := os.Open(path + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		parseFileNode(f)
	}

	path = minetest.Path("itemdefs/")
	os.Mkdir(path, 0777)

	dir, err = os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range dir {
		f, err := os.Open(path + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		parseFileItem(f)
	}
}

func parseFileNode(r io.Reader) {
	d := json.NewDecoder(r)

	var defs []*mt.NodeDef

	err := d.Decode(&defs)
	if err != nil {
		log.Printf("Error parsing nodedef '%s'", err.Error())
	}

	minetest.AddNodeDef(defs...)
}

func parseFileItem(r io.Reader) {
	d := json.NewDecoder(r)

	var defs []*mt.ItemDef

	err := d.Decode(&defs)
	if err != nil {
		log.Printf("Error parsing itemdef '%s'", err.Error())
	}

	var rdefs = make([]mt.ItemDef, len(defs))

	for k, v := range defs {
		rdefs[k] = *v
	}


	minetest.AddItemDef(rdefs...)
}
