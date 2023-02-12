package playerAO

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"time"
)

func GetAOID(c *minetest.Client) (mt.AOID, bool) {
	dat, ok := c.GetData("aoid")
	if !ok {
		return 0, false
	}

	id, ok := dat.(mt.AOID)
	if ok {
		return id, true
	} else {
		log.Fatalf("ClientData has unexpected Type expected %T got %T!\n", mt.AOID(0), dat)
	}

	return 0, false
}

func init() {
	chat.RegisterChatCmd("pos", func(c *minetest.Client, _ []string) {
		pp := minetest.GetPos(c)
		ppos := pp.Pos
		pos := ppos.Pos()

		chat.SendMsgf(c, mt.SysMsg, "Your position: (%f, %f, %f) pitch: %f, yaw: %f",
			pos[0], pos[1], pos[2],
			ppos.Pitch(), ppos.Yaw(),
		)
	})

	minetest.RegisterLeaveHook(func(l *minetest.Leave) {
		go func() {
			id, ok := GetAOID(l.Client)
			if ok {
				ao.RmAO(id)
			}
		}()
	})

	minetest.RegisterPosUpdater(func(clt *minetest.Client, p *minetest.ClientPos, dt time.Duration) {
		id, ok := GetAOID(clt)

		if !ok || id == 0 {
			return
		}

		// TODO: make ao_mgr/ao deal with positions so you just have to say: ao.UpdatePos(id, pos)
		a := ao.GetAO(id)
		if a == nil {
			return
		}

		ppos := p.Pos

		a.SetPos(mt.AOPos{
			Pos: ppos.Pos(),
			Rot: mt.Vec{0, ppos.Yaw()},
			Vel: ppos.Vel(),

			Interpolate: true,
		})
		a.SetBonePos("Head_Control", mt.AOBonePos{
			Pos: mt.Vec{0, 6.3, 0},
			Rot: mt.Vec{-ppos.Pitch(), 0, 0},
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
