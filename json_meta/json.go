package json_nodemeta

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/minetest/log"
	"io"
	"os"

	"encoding/json"
)

func init() {
	path := minetest.Path("nodedefs/")
	os.Mkdir(path, 0777)

	dir, err := os.ReadDir(path)
	if err != nil {
		log.Errorf("Error reading dir nodedefs/: %s", err)
		os.Exit(1)
	}

	for _, file := range dir {
		f, err := os.Open(path + file.Name())
		if err != nil {
			log.Errorf("Error opening file %s: %s", file.Name(), err)
			os.Exit(1)
		}

		n := parseFileNode(f)
		log.Printf("Added %d NodeDefs from %s\n", n, file.Name())
	}

	// ItemDefs
	path = minetest.Path("itemdefs/")
	os.Mkdir(path, 0777)

	dir, err = os.ReadDir(path)
	if err != nil {
		log.Errorf("Error reading dir itemdefs/: %s", err)
		os.Exit(1)
	}

	for _, file := range dir {
		f, err := os.Open(path + file.Name())
		if err != nil {
			log.Errorf("Error opening file %s: %s", file.Name(), err)
			os.Exit(1)
		}

		n := parseFileItem(f)
		log.Printf("Added %d ItemDefs from %s\n", n, file.Name())
	}
}

func parseFileNode(r io.Reader) int {
	var d = json.NewDecoder(r)
	var defs []*mt.NodeDef

	err := d.Decode(&defs)
	if err != nil {
		log.Printf("Error parsing nodedef '%s'", err.Error())
	}

	for _, def := range defs {
		minetest.AddNodeDef(minetest.NodeDef{NodeDef: *def})
	}

	return len(defs)
}

func parseFileItem(r io.Reader) int {
	var d = json.NewDecoder(r)
	var defs []*mt.ItemDef

	err := d.Decode(&defs)
	if err != nil {
		log.Printf("Error parsing itemdef '%s'", err.Error())
	}

	var rdefs []minetest.ItemDef
	for _, v := range defs {
		def, ok := minetest.TryItemDef(*v)
		if !ok {
			log.Printf("[WARN] item '%s' cant be converted into minetest.ItemDef!\n", v.Name)

			continue
		}

		rdefs = append(rdefs, def)
	}

	minetest.AddItemDef(rdefs...)

	return len(defs)
}
