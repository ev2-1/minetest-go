package playerAO

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"

	"time"
)

func init() {
	chat.RegisterChatCmd("pos", func(c *minetest.Client, _ []string) {
		pos := minetest.GetPos(c)

		chat.SendMsgf(c, mt.SysMsg, "Your position: (%f, %f, %f, dim: %s (%d)) pitch: %f, yaw: %f",
			pos.Pos[0], pos.Pos[1], pos.Pos[2], pos.Dim.String(), pos.Dim,
			pos.Pitch, pos.Yaw,
		)
	})

	minetest.RegisterLeaveHook(func(l *minetest.Leave) {
		go func() {
			id, ok := ao.GetCltAOID(l.Client)
			if ok {
				ao.RmAO(id)
			}
		}()
	})

	minetest.RegisterPosUpdater(func(clt *minetest.Client, p *minetest.ClientPos, dt time.Duration) {
		id, ok := ao.GetCltAOID(clt)

		if !ok || id == 0 {
			return
		}

		// TODO: make ao_mgr/ao deal with positions so you just have to say: ao.UpdatePos(id, pos)
		a := ao.GetAO(id)
		if a == nil {
			return
		}

		ppos := p.Pos

		a.SetPos(ao.AOPos{
			Pos: ppos,
			Rot: mt.Vec{0, ppos.Yaw},
			Vel: ppos.Vel,
		}, true)

		a.SetBonePos("Head_Control", mt.AOBonePos{
			Pos: mt.Vec{0, 6.3, 0},
			Rot: mt.Vec{-ppos.Pitch, 0, 0},
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
