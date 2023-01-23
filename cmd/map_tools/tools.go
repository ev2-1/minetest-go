package main

import (
	"github.com/EliasFleckenstein03/mtmap"
	"github.com/anon55555/mt"
	_ "github.com/mattn/go-sqlite3" // MIT licensed.

	"bytes"
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: map_tools <map.sqlite> <operation> <pos> [options]...\nOperations := { SETMETA, GETMETA, SETINV, GETINV }")
	}
	var (
		dbFile  = os.Args[1]
		op      = os.Args[2]
		posA    = os.Args[3]
		options = os.Args[4:]
	)

	posS := strings.SplitN(posA, ",", 4)
	if len(posS) != 3 {
		log.Fatal("This is 3d space, so ur pos is wrong!\n")
	}

	var pos [3]int16
	for k := range pos {
		tmp, err := strconv.ParseInt(posS[k], 10, 16)
		if err != nil {
			log.Fatalf("POS: ERR: %s\n", err)
		}

		pos[k] = int16(tmp)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatalf("Cant open '%s': %s", dbFile, err)
	}

	read, err := db.Prepare("SELECT data FROM blocks WHERE pos = ?")
	if err != nil {
		log.Fatal("Cant prepare reading statement: ", err)
	}

	write, err := db.Prepare("INSERT OR REPLACE INTO blocks (pos, data) VALUES (?, ?)")
	if err != nil {
		log.Fatal("Cant prepare writing statement: ", err)
	}
	_ = write
	switch op {
	case "SETMETA":
		if len(options) < 2 || options[0] == "help" {
			log.Fatal("Usage: SETMETA <field> <data>...")
		}

		blkpos, ii := mt.Pos2Blkpos(pos)
		dbpos := Blk2DBPos(blkpos)
		log.Printf("Pos: %d %d %d", pos[0], pos[1], pos[2])
		log.Printf("Blkpos: %d %d %d - %d\n", blkpos[0], blkpos[1], blkpos[2], ii)
		log.Printf("DBPos: %d\n", dbpos)

		r := read.QueryRow(dbpos)

		var buf []byte
		r.Scan(&buf)
		if len(buf) == 0 {
			log.Fatal("len(buf) == 0")
		}

		reader := bytes.NewReader(buf)

		blk := mtmap.Deserialize(reader, map[string]mt.Content{})
		_, ok := blk.NodeMetas[ii]
		if !ok {
			blk.NodeMetas[ii] = &mt.NodeMeta{}
		}

		var set bool
		if blk.NodeMetas[ii] == nil {
			blk.NodeMetas = map[uint16]*mt.NodeMeta{}
		}

		if blk.NodeMetas[ii].Fields == nil {
			blk.NodeMetas[ii].Fields = []mt.NodeMetaField{}
		}

		if !set {
			for i := 0; i < len(blk.NodeMetas[ii].Fields); i++ {
				if blk.NodeMetas[ii].Fields[i].Name == options[0] {
					blk.NodeMetas[ii].Fields[i].Value = strings.Join(options[1:], " ")

					set = true
				}
			}
		}

		if !set {
			blk.NodeMetas[ii].Fields = append(blk.NodeMetas[ii].Fields, mt.NodeMetaField{
				Field: mt.Field{
					Name:  options[0],
					Value: strings.Join(options[1:], " "),
				},
			})
		}

		b := &bytes.Buffer{}
		mtmap.Serialize(blk, b, map[mt.Content]string{})

		_, err := write.Exec(dbpos, b.Bytes())
		if err != nil {
			log.Fatalf("Err: %s\n", err)
		}

		log.Printf("Done!")

	case "SETINV":
		if len(options) < 1 || options[0] == "help" {
			log.Fatal("Usage: SETINV <data>...")
		}

		blkpos, ii := mt.Pos2Blkpos(pos)
		dbpos := Blk2DBPos(blkpos)
		log.Printf("Pos: %d %d %d", pos[0], pos[1], pos[2])
		log.Printf("Blkpos: %d %d %d - %d\n", blkpos[0], blkpos[1], blkpos[2], ii)
		log.Printf("DBPos: %d\n", dbpos)

		// serialize inv:
		inv := new(mt.Inv)
		sreader := strings.NewReader(strings.Join(options, " "))
		inv.Deserialize(sreader)

		r := read.QueryRow(dbpos)

		var buf []byte
		r.Scan(&buf)
		if len(buf) == 0 {
			log.Fatal("len(buf) == 0")
		}

		reader := bytes.NewReader(buf)

		blk := mtmap.Deserialize(reader, map[string]mt.Content{})
		_, ok := blk.NodeMetas[ii]
		if !ok {
			blk.NodeMetas[ii] = &mt.NodeMeta{}
		}

		if blk.NodeMetas[ii] == nil {
			blk.NodeMetas = map[uint16]*mt.NodeMeta{}
		}

		blk.NodeMetas[ii].Inv = *inv

		b := &bytes.Buffer{}
		mtmap.Serialize(blk, b, map[mt.Content]string{})

		_, err := write.Exec(dbpos, b.Bytes())
		if err != nil {
			log.Fatalf("Err: %s\n", err)
		}

		log.Printf("Done!")

	case "GETMETA":
		if len(options) < 1 || options[0] == "help" {
			log.Fatal("Usage: GETMETA [field]")
		}

		var field string
		if len(options) >= 2 {
			field = options[0]
		}

		blkpos, i := mt.Pos2Blkpos(pos)
		dbpos := Blk2DBPos(blkpos)
		log.Printf("Pos: %d %d %d", pos[0], pos[1], pos[2])
		log.Printf("Blkpos: %d %d %d - %d\n", blkpos[0], blkpos[1], blkpos[2], i)
		log.Printf("DBPos: %d\n", dbpos)

		r := read.QueryRow(dbpos)

		var buf []byte
		r.Scan(&buf)
		if len(buf) == 0 {
			log.Fatal("len(buf) == 0")
		}

		reader := bytes.NewReader(buf)

		blk := mtmap.Deserialize(reader, map[string]mt.Content{})
		meta, ok := blk.NodeMetas[i]
		if !ok {
			log.Printf("Meta: N/A")
		} else {
			if field == "" {

				log.Printf("Meta: %#v", meta)
			} else {
				for _, v := range meta.Fields {
					if v.Name == field {
						log.Printf("'%s': %s\n", field, v.Value)
					}
				}
			}
		}

	case "GETINV":
		blkpos, i := mt.Pos2Blkpos(pos)
		dbpos := Blk2DBPos(blkpos)
		log.Printf("Pos: %d %d %d", pos[0], pos[1], pos[2])
		log.Printf("Blkpos: %d %d %d - %d\n", blkpos[0], blkpos[1], blkpos[2], i)
		log.Printf("DBPos: %d\n", dbpos)

		r := read.QueryRow(dbpos)

		var buf []byte
		r.Scan(&buf)
		if len(buf) == 0 {
			log.Fatal("len(buf) == 0")
		}

		reader := bytes.NewReader(buf)

		blk := mtmap.Deserialize(reader, map[string]mt.Content{})
		meta, ok := blk.NodeMetas[i]
		if !ok {
			log.Printf("Meta: N/A")
		} else {
			log.Printf("Meta: %#v", meta.Inv)
		}
	}
}

func Blk2DBPos(p [3]int16) int64 {
	return int64(p[2])*16777216 + int64(p[1])*4096 + int64(p[0])
}
