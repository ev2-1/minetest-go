package playerAO

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/tools/pos"

	"log"
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
		pp := pos.GetPos(c)
		pos := pp.Pos()

		chat.SendMsgf(c, mt.SysMsg, "Your position: (%f, %f, %f) pitch: %f, yaw: %f",
			pos[0], pos[1], pos[2],
			pp.Pitch(), pp.Yaw(),
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

	pos.RegisterPosUpdater(func(clt *minetest.Client, p mt.PlayerPos, dt int64) {
		id, ok := GetAOID(clt)

		if !ok || id == 0 {
			return
		}

		// TODO: make ao_mgr/ao deal with positions so you just have to say: ao.UpdatePos(id, pos)
		a := ao.GetAO(id)
		if a == nil {
			return
		}

		a.SetPos(mt.AOPos{
			Pos: p.Pos(),
			Rot: mt.Vec{0, p.Yaw()},
			Vel: p.Vel(),

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
