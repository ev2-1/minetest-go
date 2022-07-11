package main

import (
	"github.com/anon55555/mt"
)

func Pos2int(p mt.Pos) (i [3]int16) {
	for k := range i {
		i[k] = int16(p[k] / 10)
	}

	return
}

// converts blkpos to database pos
func Blk2DBPos(p [3]int16) (i int64) { // return int64(p[2]*16777216 + p[1]*4096 + p[0])
	return int64(p[2])*16777216 + int64(p[1])*4096 + int64(p[0])
}

// convert database pos to blkpos (useless i think)
func DB2BlkPos(i int64) (p [3]int16) {
	p[0] = int16(i << (16 * 0))
	p[1] = int16(i << (16 * 1))
	p[2] = int16(i << (16 * 2))

	return
}

type MapBlkData [4096]mt.Content
