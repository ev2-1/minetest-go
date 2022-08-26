package playerAO

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/tools/pos"

	"fmt"
	"sync"
	"time"
)

var clients = make(map[*minetest.Client]mt.AOID)
var clientsMu sync.RWMutex

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvChatMsg:
			switch cmd.Msg { // return own pos
			case "pos":
				pp := pos.GetPos(c)
				pos := pp.Pos()

				c.SendCmd(&mt.ToCltChatMsg{
					Type: mt.RawMsg,

					Text: fmt.Sprintf("Your position: (%f, %f, %f) pitch: %f, yaw: %f",
						pos[0], pos[1], pos[2],
						pp.Pitch(), pp.Yaw(),
					),

					Timestamp: time.Now().Unix(),
				})
			}
		}
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

	pos.RegisterPosUpdater(func(clt *minetest.Client, p *mt.PlayerPos, dt int64) {
		clientsMu.RLock()
		id, ok := clients[clt]
		clientsMu.RUnlock()
		if !ok {
			return
		}

		// TODO: make ao_mgr/ao deal with positions so you just have to say: ao.UpdatePos(id, pos)
		ao.AOMsg(
			mt.IDAOMsg{
				ID: id,
				Msg: &mt.AOCmdPos{
					Pos: mt.AOPos{
						Pos: p.Pos(),
						Rot: mt.Vec{0, p.Yaw()},

						Interpolate: true,
					},
				},
			},
			mt.IDAOMsg{
				ID: id,
				Msg: &mt.AOCmdBonePos{
					Bone: "Head_Control",
					Pos: mt.AOBonePos{
						Pos: mt.Vec{0, 6.3, 0},
						Rot: mt.Vec{-p.Pitch(), 0, 0},
					},
				},
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
