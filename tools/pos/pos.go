package pos

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"sync"
	"time"
)

type ClientPos struct {
	sync.RWMutex

	Pos        mt.PlayerPos
	OldPos     mt.PlayerPos
	LastUpdate time.Time
}

var posUpdatersMu sync.RWMutex
var posUpdaters []func(c *minetest.Client, pos *ClientPos, lu time.Duration)

// PosUpdater is called with a LOCKED ClientPos
func RegisterPosUpdater(pu func(c *minetest.Client, pos *ClientPos, lu time.Duration)) {
	posUpdatersMu.Lock()
	defer posUpdatersMu.Unlock()

	posUpdaters = append(posUpdaters, pu)
}

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		c.Logf("Pkt: %T", pkt.Cmd)
	})

	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		pp, ok := pkt.Cmd.(*mt.ToSrvPlayerPos)

		if ok {
			cpos := GetPos(c)
			cpos.Lock()
			defer cpos.Unlock()

			now := time.Now()
			dtime := now.Sub(cpos.LastUpdate)

			cpos.LastUpdate = now
			cpos.OldPos = cpos.Pos
			cpos.Pos = pp.Pos

			c.Logf("start u")
			for _, u := range posUpdaters {
				u(c, cpos, dtime)
			}
			c.Logf("done u")
		}
	})
}

func MakePos(c *minetest.Client) *ClientPos {
	return &ClientPos{
		Pos:        mt.PlayerPos{Pos100: [3]int32{0, 100, 100}},
		LastUpdate: time.Now(),
	}
}

// GetPos returns pos os player / client
func GetPos(c *minetest.Client) *ClientPos {
	cd, ok := c.GetData("pos")
	if !ok {
		cd = MakePos(c)
		c.SetData("pos", cd)
	}

	pos, ok := cd.(*ClientPos)
	if !ok {
		pos = MakePos(c)
		c.SetData("pos", cd)
	} else {
	}
	return pos
}

// SetPos sets position
// returns old position
func SetPos(c *minetest.Client, p mt.PlayerPos) mt.PlayerPos {
	cpos := GetPos(c)
	cpos.Lock()
	defer cpos.Unlock()

	cpos.OldPos = cpos.Pos
	cpos.Pos = p

	return cpos.OldPos
}
