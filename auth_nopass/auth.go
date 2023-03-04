package auth_nopass

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"time"
)

func Stage1() {}
func Stage2() {}

func init() {
	minetest.RegisterRawPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvInit:
			if c.State > minetest.CsCreated {
				c.Log("->", "duplicate init")

				return
			}

			c.SetState(minetest.CsInit)

			if cmd.SerializeVer < minetest.SerializeVer {
				c.Log("<-", "invalid serializeVer", cmd.SerializeVer)
				minetest.CltLeave(&minetest.Leave{
					Reason: mt.UnsupportedVer,

					Client: c,
				})

				return
			}

			if cmd.MaxProtoVer < minetest.ProtoVer {
				c.Log("<-", "invalid protoVer", cmd.MaxProtoVer)
				minetest.CltLeave(&minetest.Leave{
					Reason: mt.UnsupportedVer,

					Client: c,
				})

				return
			}

			if len(cmd.PlayerName) == 0 || len(cmd.PlayerName) > minetest.MaxPlayerNameLen {
				c.Log("<-", "invalid player name length")
				minetest.CltLeave(&minetest.Leave{
					Reason: mt.BadName,

					Client: c,
				})

				return
			}
			c.Name = cmd.PlayerName

			if minetest.PlayerExists(c.Name) {
				c.Log("<-", "player already joined")
				minetest.CltLeave(&minetest.Leave{
					Reason: mt.AlreadyConnected,

					Client: c,
				})

				return
			}

			// reply is always FirstSRP
			c.Log("send to clt hello")
			c.SendCmd(&mt.ToCltHello{
				SerializeVer: minetest.SerializeVer,
				ProtoVer:     minetest.ProtoVer,
				AuthMethods:  mt.FirstSRP,
				Username:     c.Name,
			})

		case *mt.ToSrvFirstSRP:
			minetest.InitClient(c)
			cpos := c.GetPos()

			c.SendCmd(&mt.ToCltAcceptAuth{
				PlayerPos: mt.Pos{
					cpos.Pos.Pos[0],
					cpos.Pos.Pos[1] + 5,
					cpos.Pos.Pos[2],
				},
				MapSeed:         1337,
				SendInterval:    0.09,
				SudoAuthMethods: mt.SRP,
			})

			c.SendCmd(&mt.ToCltMovePlayer{
				Pos: cpos.Pos.Pos,

				Pitch: cpos.Pitch,
				Yaw:   cpos.Yaw,
			})

			c.SetState(minetest.CsActive)

		case *mt.ToSrvInit2:
			c.SendItemDefs()
			c.SendNodeDefs()
			c.SendAnnounceMedia()

			// is ignored anyways
			c.SendCmd(&mt.ToCltCSMRestrictionFlags{MapRange: 3})

		case *mt.ToSrvCltReady:
			if c.State == minetest.CsActive {
				// Wait to ensure client is *actually* ready (TODO: find better work around)
				time.Sleep(time.Second)

				minetest.RegisterPlayer(c)
			} else {
				minetest.CltLeave(&minetest.Leave{
					Reason: mt.UnexpectedData,

					Client: c,
				})
			}

			return

		case *mt.ToSrvGotBlks:
			return

		default:
			return
		}

		return
	})
}
