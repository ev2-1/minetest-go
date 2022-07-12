package main

import (
	"database/sql"
	"github.com/anon55555/mt"
	_ "github.com/mattn/go-sqlite3" // MIT licensed.
	"log"
)

// DecodeBlkBlob parses a blkblob from a database to mapblkdata
func DecodeBlkBlob(data DBResult) (blk mt.MapBlk, ok bool) {
	if len(data.param0) != 4096*2 {
		return mt.MapBlk{}, false
	}

	for k := range blk.Param0 {
		blk.Param0[k] = mt.Content(data.param0[k*2])<<8 | mt.Content(data.param0[k*2+1])<<0
	}

	return blk, true
}

// EncodeBlkBlob encodes mapblkdata to blkblob from a database
func EncodeBlkBlob(blk *mt.MapBlk) (data *DBResult) {
	data = &DBResult{}

	if blk == nil {
		return
	}

	arr := [4096 * 2]byte{}
	data.param0 = arr[:]

	for k := range blk.Param0 {
		data.param0[k*2] = byte(blk.Param0[k] >> 8)
		data.param0[k*2+1] = byte(blk.Param0[k])
	}

	return
}

var db *sql.DB

var writeBlk *sql.Stmt
var readBlk *sql.Stmt

func OpenDB(file string) (err error) {
	db, err = sql.Open("sqlite3", file)
	if err != nil {
		log.Fatal("cant open map.sqlite: ", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS \"blocks\" ( \"pos\" INTEGER, \"param0\" BLOB, \"param1\" BLOB, \"param2\" BLOB, PRIMARY KEY(\"pos\") )")
	if err != nil {
		log.Fatal("cant create table blocks: ", err)
	}

	// prepare stms:
	// writeBlk, err = db.Prepare("UPSERT INTO blocks (pos, param0, param1, param2) VALUES (?, ?, ?, ?)")
	readBlk, err = db.Prepare("SELECT param0, param1, param2 FROM blocks WHERE pos = ?")
	if err != nil {
		log.Fatal("cant prepare read statement: ", err)
	}

	// return the return value to the caller
	// this is important so the caller can check whether
	// an error has occured during the execution of the function
	return
}

func GetBlk(p [3]int16) (mt.MapBlk, bool) {
	r := readBlk.QueryRow(Blk2DBPos(p))

	param0 := []byte{}
	param1 := []byte{}
	param2 := []byte{}

	r.Scan(&param0, &param1, &param2)

	data := DBResult{param0, param1, param2}

	return DecodeBlkBlob(data)
}

func SetNode(pos [3]int16, node mt.Content) {
	blk, i := mt.Pos2Blkpos(pos)
	oldBlk, ok := GetBlk(blk)

	if !ok {
		oldBlk = EmptyBlk
	}

	oldBlk.Param0[i] = node

	SetBlk(blk, &oldBlk)
}

func SetBlk(p [3]int16, blk *mt.MapBlk) {
	q, err := db.Prepare("INSERT OR REPLACE INTO blocks (pos, param0, param1, param2) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal("can't set block: ", err)
	}

	defer q.Close()

	data := EncodeBlkBlob(blk)

	if data == nil {
		log.Fatal("blob is nil")
	}

	pos := Blk2DBPos(p)

	_, err = q.Exec(pos, data.param0[:], data.param1[:], data.param2[:])
	if err != nil {
		log.Fatal(err)
	}
}

func Commit() {
}

type DBResult struct {
	param0 []byte
	param1 []uint8
	param2 []uint8
}
