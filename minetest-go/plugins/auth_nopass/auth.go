package main

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"

	"reflect"
	"fmt"
)

func init() {
	minetest.RegisterPacketHandler(&minetest.PacketHandler{
		Packets: map[reflect.Type]bool{
			reflect.TypeOf(&mt.ToSrvInit{}):      true,
			reflect.TypeOf(&mt.ToSrvInit2{}):     true,
			reflect.TypeOf(&mt.ToSrvFirstSRP{}):  true,
			reflect.TypeOf(&mt.ToSrvSRPBytesA{}): true,
			reflect.TypeOf(&mt.ToSrvSRPBytesM{}): true,
		},

		Handle: func(c *minetest.Client, pkt *mt.Pkt) bool {
			switch cmd := pkt.Cmd.(type) {
			case *mt.ToSrvInit:
				if c.State > minetest.CsCreated {
					c.Log("->", "duplicate init")
				}

				c.SetState(minetest.CsInit)
				
				if cmd.SerializeVer < minetest.SerializeVer {
					c.Log("<-", "invalid serializeVer", cmd.SerializeVer)
					minetest.CltLeave(&minetest.Leave{
						Reason: mt.UnsupportedVer,

						Client: c,
					})
				}

				if cmd.MaxProtoVer < minetest.ProtoVer {
					c.Log("<-", "invalid protoVer", cmd.MaxProtoVer)
					minetest.CltLeave(&minetest.Leave{
						Reason: mt.UnsupportedVer,

						Client: c,
					})
				}

				if len(cmd.PlayerName) == 0 || len(cmd.PlayerName) > minetest.MaxPlayerNameLen {
					c.Log("<-", "invalid player name length")
					minetest.CltLeave(&minetest.Leave{
						Reason: mt.BadName,

						Client: c,
					})					
				}

				c.Name = cmd.PlayerName
				c.Logger.SetPrefix(fmt.Sprintf(""))

				if minetest.PlayerExists(c.Name) {
					minetest.CltLeave(&minetest.Leave{
						Reason: mt.AlreadyConnected,

						Client: c,
					})
				}

				minetest.RegisterPlayer(c)

				// reply is always FirstSRP
				c.SendCmd(&mt.ToCltHello{
					SerializeVer: minetest.SerializeVer,
					ProtoVer:     minetest.ProtoVer,
					AuthMethods:  mt.FirstSRP,
					Username:     c.Name,
				})

			case *mt.ToSrvFirstSRP:
				c.SendCmd(&mt.ToCltAcceptAuth{
					PlayerPos:       mt.Pos{0, 100, 0},
					MapSeed:         1337,
					SendInterval:    0.09,
					SudoAuthMethods: mt.SRP,
				})

				c.SetState(minetest.CsActive)

			case *mt.ToSrvInit2:
				c.SendItemDefs()
				c.SendNodeDefs()
				c.SendAnnounceMedia()

				// is ignored anyways			
				c.SendCmd(&mt.ToCltCSMRestrictionFlags{})
				
			default:
				return false
			}

			return true
		},
	})
}
