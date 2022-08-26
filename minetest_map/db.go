package minetest_map

import (
	"bytes"
	"database/sql"
	"github.com/EliasFleckenstein03/mtmap"
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	_ "github.com/mattn/go-sqlite3" // MIT licensed.

	"log"
	"sync"
)

var db *sql.DB

var dbMu sync.Mutex

var writeBlk *sql.Stmt
var readBlk *sql.Stmt

func OpenDB(file string) (err error) {
	db, err = sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal("cant open map.sqlite: ", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `blocks` (	`pos` INT PRIMARY KEY, `data` BLOB );")
	if err != nil {
		log.Fatal("cant create table blocks: ", err)
	}

	// prepare stms:
	// writeBlk, err = db.Prepare("UPSERT INTO blocks (pos, param0, param1, param2) VALUES (?, ?, ?, ?)")
	readBlk, err = db.Prepare("SELECT data FROM blocks WHERE pos = ?")
	if err != nil {
		log.Fatal("cant prepare read statement: ", err)
	}

	// return the return value to the caller
	// this is important so the caller can check whether
	// an error has occured during the execution of the function
	return
}

func GetBlk(p [3]int16) <-chan *mtmap.MapBlk {
	r := readBlk.QueryRow(Blk2DBPos(p))

	var buf []byte
	r.Scan(&buf)
	if len(buf) == 0 {
		return nil
	}

	ch := make(chan *mtmap.MapBlk)

	go func() {
		reader := bytes.NewReader(buf)

		ch <- mtmap.Deserialize(reader, minetest.IdNodeMap)
		close(ch)
	}()

	return ch
}

func SetBlk(p [3]int16, blk *mtmap.MapBlk) {
	q, err := db.Prepare("INSERT OR REPLACE INTO blocks (pos, data) VALUES (?, ?)")
	if err != nil {
		log.Fatal("can't set block: ", err)
	}

	defer q.Close()

	w := &bytes.Buffer{}

	mtmap.Serialize(blk, w, minetest.NodeIdMap)

	pos := Blk2DBPos(p)

	_, err = q.Exec(pos, w.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}

// EmptyBlk returns a empty MapBlock containing fully lit air
func EmptyBlk() (blk *mtmap.MapBlk) {
	for k := range blk.Param0 {
		blk.Param0[k] = mt.Air
		blk.Param1[k] = 255
	}

	return
}
