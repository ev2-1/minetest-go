package minetest_map

import (
	"bytes"
	"database/sql"
	"github.com/EliasFleckenstein03/mtmap"
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	_ "github.com/mattn/go-sqlite3" // MIT licensed.

	"errors"
	"fmt"
	"sync"
	"time"
)

/*
type MapDriver interface {
	Open(string) error

	GetBlk([3]int16) MapBlk
	SetBlk([3]int16, MapBlk)
}
*/

type MinetestMapDriver struct {
	db   *sql.DB
	dbMu sync.Mutex

	read  *sql.Stmt
	write *sql.Stmt

	NodeIdMap map[mt.Content]string
	IdNodeMap map[string]mt.Content
}

func init() {
	minetest.RegisterMapDriver("minetest", &MinetestMapDriver{})
}

var (
	ErrDBMapBlockNoData = errors.New("No Data in field BLOB")
)

func (drv *MinetestMapDriver) Make() minetest.MapDriver {
	return new(MinetestMapDriver)
}

func (drv *MinetestMapDriver) Open(file string) error {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return fmt.Errorf("Can't open '%s': %s", file, err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS `blocks` (	`pos` INT PRIMARY KEY, `data` BLOB );")
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Cant create Table blocks in file '%s':", file), err)
	}

	// prepare Stmts
	drv.read, err = db.Prepare("SELECT data FROM blocks WHERE pos = ?")
	if err != nil {
		return fmt.Errorf("Cant prepare reading statement: %s", err)
	}

	drv.write, err = db.Prepare("INSERT OR REPLACE INTO blocks (pos, data) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("Cant prepare writing statement: %s", err)
	}

	drv.NodeIdMap, drv.IdNodeMap = minetest.NodeMaps()

	return nil
}

func Blk2DBPos(p [3]int16) int64 {
	return int64(p[2])*16777216 + int64(p[1])*4096 + int64(p[0])
}

func (drv *MinetestMapDriver) GetBlk(pos [3]int16) (*minetest.MapBlk, error) {
	blk, err := drv.readBlkFromDB(pos)
	if err != nil {
		return nil, err
	}

	bblk := &minetest.MapBlk{
		MapBlk: blk.MapBlk,
		Pos:    pos,

		Driver: drv,

		Loaded: time.Now(),
	}

	return bblk, nil
}

func (drv *MinetestMapDriver) SetBlk(blk *minetest.MapBlk) error {
	drv.writeBlkToDB(blk.Pos, &mtmap.MapBlk{
		MapBlk:    blk.MapBlk,
		Timestamp: uint32(blk.LastAccess.Unix()),
	})

	return nil
}

func (drv *MinetestMapDriver) readBlkFromDB(p [3]int16) (*mtmap.MapBlk, error) {
	r := drv.read.QueryRow(Blk2DBPos(p))

	var buf []byte
	r.Scan(&buf)
	if len(buf) == 0 {
		return nil, ErrDBMapBlockNoData
	}

	reader := bytes.NewReader(buf)

	return mtmap.Deserialize(reader, drv.IdNodeMap), nil
}

func (drv *MinetestMapDriver) writeBlkToDB(p [3]int16, blk *mtmap.MapBlk) error {
	w := &bytes.Buffer{}

	mtmap.Serialize(blk, w, drv.NodeIdMap)

	pos := Blk2DBPos(p)

	_, err := drv.write.Exec(pos, w.Bytes())
	if err != nil {
		return err
	}

	return nil
}
