package playerAO

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"
)

func init() {
	chat.RegisterChatCmd("pos", func(c *minetest.Client, _ []string) {
		pos := minetest.GetPos(c)

		chat.SendMsgf(c, mt.SysMsg, "Your position: [%#v]",
			pos,
		)
	})

	/*
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
	*/
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
