package tnt_ao

import (
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"

	"github.com/anon55555/mt"

	"image/color"
	"strconv"
)

func init() {
	chat.RegisterChatCmd("spawn_tnt", func(c *minetest.Client, args []string) {
		ao.RegisterAO(testAO(ao.AOPos{Pos: minetest.GetPos(c)}))
	})

	chat.RegisterChatCmd("rm_ao", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "usage: rm_ao <aoid (uint16)>", mt.SysMsg)
			return
		}

		id, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			chat.SendMsgf(c, mt.SysMsg, "invalid aoid \"%s\"", args[0])
			return
		}

		ao.RmAO(mt.AOID(id))
	})
}

func testAO(pos ao.AOPos) ao.ActiveObject {
	return &ao.ActiveObjectS{
		AOState: ao.AOState{
			Pos: pos,
			HP:  10,
		},

		Props: mt.AOProps{
			Mesh:           "",
			MaxHP:          10,
			Pointable:      false,
			CollideWithAOs: true,
			ColBox: mt.Box{
				mt.Vec{-0.5, -0.5, -0.5},
				mt.Vec{0.5, 0.5, 0.5},
			},
			SelBox: mt.Box{
				mt.Vec{-0.5, -0.5, -0.5},
				mt.Vec{0.5, 0.5, 0.5},
			},
			Visual:          "cube",
			VisualSize:      [3]float32{1.0, 1.0, 1.0},
			Textures:        []mt.Texture{"default_tnt_top.png", "default_tnt_bottom.png", "default_tnt_side.png", "default_tnt_side.png", "default_tnt_side.png", "default_tnt_side.png"},
			DmgTextureMod:   "^[brighten",
			Shaded:          true,
			SpriteSheetSize: [2]int16{1, 1},
			SpritePos:       [2]int16{0, 0},
			Visible:         true,
			Colors:          []color.NRGBA{color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}},
			BackfaceCull:    true,
			NametagColor:    color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
			NametagBG:       color.NRGBA{R: 0x01, G: 0x01, B: 0x01, A: 0x00},
			FaceRotateSpeed: -1,
			Infotext:        "",
			Itemstring:      "",
		},
	}
}
