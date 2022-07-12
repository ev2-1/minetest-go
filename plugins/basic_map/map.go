package main

import (
	"github.com/anon55555/mt"
	_minetest "github.com/ev2-1/minetest-go"
	"github.com/ev2-1/minetest-go/abstract"

	"log"
	"time"
)

var posCh <-chan *minetest.CltPos

// a list of all clients and their loaded chunks
var loadedChunks map[string]map[pos]int64

var joinCh <-chan *_minetest.Client
var leaveCh <-chan *_minetest.Leave

var (
	MapBlkUpdateRate  int64 = 2 // in seconds
	MapBlkUpdateRange       = 5 // in mapblks
	EmptyBlk          mt.MapBlk
)

var exampleBlk mt.MapBlk

func init() {
	posCh = minetest.GetPosCh()
	joinCh = _minetest.JoinChan()
	leaveCh = _minetest.LeaveChan()

	loadedChunks = make(map[string]map[pos]int64)

	OpenDB(_minetest.Path("/map.sqlite"))

	exampleBlk = mt.MapBlk{}

	for i := 0; i < 4096; i++ {
		exampleBlk.Param0[i] = 126
	}

	for i := 0; i < 16*16; i++ {
		exampleBlk.Param0[i] = 349
	}

	// center block is stone:
	exampleBlk.Param0[4096/2+16/2] = 349 // some wool

	for k := range EmptyBlk.Param0 {
		EmptyBlk.Param0[k] = mt.Air
	}

	go func() {
		for {
			pos, ok := <-posCh
			if !ok {
				log.Print("[ERROR]", "mapblk pos chan not ok")
				return
			}

			if time.Now().Unix() < pos.LastUpdate+MapBlkUpdateRate {
				p := Pos2int(pos.Pos())
				blkpos, _ := mt.Pos2Blkpos(p)

				name := pos.Name

				for _, sp := range spiral(int16(MapBlkUpdateRange)) {
					// generate absolute position
					ap := sp.add(blkpos)

					// load block
					blk := LoadChunk(name, ap)

					// if block has content; send to clt
					if blk != nil {
						go pos.SendCmd(&mt.ToCltBlkData{
							Blkpos: ap,
							Blk:    *blk,
						})
					}
				}
			}
		}
	}()

	go func() {
		for {
			c, ok := <-joinCh
			if !ok {
				log.Fatal("join channel broke")
			}

			loadedChunks[c.Name] = make(map[pos]int64)
		}
	}()

	go func() {
		for {
			l, ok := <-leaveCh
			if !ok {
				log.Fatal("leave channel broke")
			}

			c := l.Client
			delete(loadedChunks, c.Name)

		}
	}()

	// interactions:
	initInteractions()
}

func LoadChunk(name string, p pos) *mt.MapBlk {
	if loadedChunks[name] == nil {
		loadedChunks[name] = make(map[pos]int64)
	}

	t := time.Now().Unix()

	if !(loadedChunks[name][p] < t-MapBlkUpdateRate) {
		return nil
	}

	blkdata, ok := GetBlk(p)
	if !ok {
		SetBlk(p, &exampleBlk)
		blkdata = exampleBlk
	}

	loadedChunks[name][p] = t

	return &blkdata
}
