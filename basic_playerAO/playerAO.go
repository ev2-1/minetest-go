package playerAO

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/tools/pos"

	"sync"
)

var clients = make(map[*minetest.Client]mt.AOID)
var clientsMu sync.RWMutex

func init() {
	chat.RegisterChatCmd("pos", func(c *minetest.Client, _ []string) {
		pp := pos.GetPos(c)
		pos := pp.Pos()

		chat.SendMsgf(c, mt.SysMsg, "Your position: (%f, %f, %f) pitch: %f, yaw: %f",
			pos[0], pos[1], pos[2],
			pp.Pitch(), pp.Yaw(),
		)
	})

	minetest.RegisterLeaveHook(func(l *minetest.Leave) {
		go func() {
			clientsMu.Lock()
			defer clientsMu.Unlock()

			if _, ok := clients[l.Client]; ok {
				ao.RmAO(clients[l.Client])

				delete(clients, l.Client)
			}
		}()
	})

	pos.RegisterPosUpdater(func(clt *minetest.Client, p mt.PlayerPos, dt int64) {
		clientsMu.RLock()
		defer clientsMu.RUnlock()

		id, ok := clients[clt]

		if !ok || id == 0 {
			return
		}

		// TODO: make ao_mgr/ao deal with positions so you just have to say: ao.UpdatePos(id, pos)
		a := ao.GetAO(id)
		a.SetPos(mt.AOPos{
			Pos: p.Pos(),
			Rot: mt.Vec{0, p.Yaw()},

			Interpolate: true,
		})
		a.SetBonePos("Head_Control", mt.AOBonePos{
			Pos: mt.Vec{0, 6.3, 0},
			Rot: mt.Vec{-p.Pitch(), 0, 0},
		})
	})

	ao.RegisterAO0Maker(makeAO)
}

/*
func playerAO(c *minetest.Client, self bool) {
	var p mt.PlayerPos
	var aoid mt.AOID
	if !self {
		p = pos.GetPos(c)
		aoid = playerInitialized[c].ID
	} else {
		aoid = 0
	}

	name := c.Name

}
*/
