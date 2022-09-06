package mmap

import (
	"bytes"
	"database/sql"
	"github.com/EliasFleckenstein03/mtmap"
	"github.com/ev2-1/minetest-go/minetest"
	_ "github.com/mattn/go-sqlite3" // MIT licensed.

	"errors"
	"fmt"
	"log"
	"sync"
)

var (
	db   *sql.DB
	dbMu sync.Mutex
)

var (
	read  *sql.Stmt
	write *sql.Stmt
)

var (
	ErrDBMapBlockNoData = errors.New("No Data in field BLOB")
)

func init() {
	Open(minetest.Path("map.sqlite"))
}

func Open(file string) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cant open '%s':", file), err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `blocks` (	`pos` INT PRIMARY KEY, `data` BLOB );")
	if err != nil {
		log.Fatal(fmt.Sprintf("Cant create Table blocks in file '%s':", file), err)
	}

	// prepare Stmts
	read, err = db.Prepare("SELECT data FROM blocks WHERE pos = ?")
	if err != nil {
		log.Fatal("Cant prepare reading statement: ", err)
	}

	write, err = db.Prepare("INSERT OR REPLACE INTO blocks (pos, data) VALUES (?, ?)")
	if err != nil {
		log.Fatal("Cant prepare writing statement: ", err)
	}

	return
}

func Blk2DBPos(p [3]int16) int64 {
	return int64(p[2])*16777216 + int64(p[1])*4096 + int64(p[0])
}

func readBlkFromDB(p [3]int16) (*mtmap.MapBlk, error) {
	r := read.QueryRow(Blk2DBPos(p))

	var buf []byte
	r.Scan(&buf)
	if len(buf) == 0 {
		return nil, ErrDBMapBlockNoData
	}

	reader := bytes.NewReader(buf)

	return mtmap.Deserialize(reader, minetest.IdNodeMap), nil
}

func writeBlkToDB(p [3]int16, blk *mtmap.MapBlk) error {
	w := &bytes.Buffer{}

	mtmap.Serialize(blk, w, minetest.NodeIdMap)

	pos := Blk2DBPos(p)

	_, err := write.Exec(pos, w.Bytes())
	if err != nil {
		return err
	}

	return nil
}
