package main

import (
	"database/sql"
	"github.com/anon55555/mt"
	_ "github.com/mattn/go-sqlite3" // MIT licensed.
	"log"
)

// DecodeBlkBlob parses a blkblob from a database to mapblkdata
func DecodeBlkBlob(bytes []byte) (blk MapBlkData, ok bool) {
	if len(bytes) != 4096*2 { // *2 cuz uint16
		return MapBlkData{}, false
	}

	for k := range blk {
		blk[k] = mt.Content(bytes[k*2])<<8 | mt.Content(bytes[k*2+1])<<0
	}

	return blk, true
}

// EncodeBlkBlob encodes mapblkdata to blkblob from a database
func EncodeBlkBlob(blk *MapBlkData) (bytes *[4096 * 2]byte) {
	bytes = &[4096 * 2]byte{}

	if blk == nil {
		return
	}

	for k := range blk {
		bytes[k*2] = byte(blk[k] >> 8)
		bytes[k*2+1] = byte(blk[k])
	}

	return
}

var db *sql.DB

var writeBlk *sql.Stmt
var readBlk *sql.Stmt

func OpenDB(file string) (err error) {
	db, err = sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("CREATE TABLE IF NOT EXISTS blocks (pos INT PRIMARY KEY, data BLOB)")

	// prepare stms:
	writeBlk, err = db.Prepare("UPSERT INTO blocks (pos, data) VALUES (?, ?)")
	readBlk, err = db.Prepare("SELECT data FROM blocks WHERE pos = ?")
	if err != nil {
		log.Fatal(err)
	}

	// return the return value to the caller
	// this is important so the caller can check whether
	// an error has occured during the execution of the function
	return
}

func GetBlk(p [3]int16) (MapBlkData, bool) {
	r := readBlk.QueryRow(Blk2DBPos(p))

	data := []byte{}

	r.Scan(&data)

	return DecodeBlkBlob(data)
}

func SetNode(pos [3]int16, node mt.Content) {
	blk, i := mt.Pos2Blkpos(pos)
	oldBlk, ok := GetBlk(blk)

	if !ok {
		oldBlk = EmptyBlk
	}

	oldBlk[i] = node

	SetBlk(blk, &oldBlk)
}

func SetBlk(p [3]int16, blk *MapBlkData) {
	q, err := db.Prepare("INSERT OR REPLACE INTO blocks (pos, data) VALUES (?, ?)")
	defer q.Close()
	if err != nil {
		log.Fatal(err)
	}

	data := EncodeBlkBlob(blk)
	if data == nil {
		log.Fatal("blob is nil")
	}

	pos := Blk2DBPos(p)

	_, err = q.Exec(pos, data[:])
	if err != nil {
		log.Fatal(err)
	}
}

func Commit() {
}
