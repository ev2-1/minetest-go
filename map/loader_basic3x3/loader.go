package mapLoader

import (
	"github.com/ev2-1/minetest-go/minetest"
	"log"
	"sync"
)

type Loader3x3 struct {
	sync.Mutex

	clt *minetest.Client

	lastDim *minetest.DimID
}

func (l *Loader3x3) Load() {
	l.Lock()
	defer l.Unlock()

	if l.clt == nil {
		log.Fatalf("Loader3x3.Load() called without clt.")
	}

	pos := l.clt.GetFullPos().Copy()

	pos.RLock()
	newPos, _ := minetest.Pos2Blkpos(pos.CurPos.IntPos())
	oldPos, _ := minetest.Pos2Blkpos(pos.OldPos.IntPos())
	pos.RUnlock()

	if newPos != oldPos || l.lastDim == nil || pos.CurPos.Dim != *l.lastDim {
		go loadAround(newPos, l.clt)
	}
}

func (l *Loader3x3) Make(clt *minetest.Client) minetest.MapLoader {
	return &Loader3x3{clt: clt}
}

func init() {
	minetest.RegisterMapLoader("loader3x3", &Loader3x3{})
}
