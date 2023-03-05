package tnt_ao

import (
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/minetest/log"

	"github.com/anon55555/mt"

	"image/color"
	"strconv"
	"sync"
	"time"
)

func init() {
	chat.RegisterChatCmd("spawn_tnt", func(c *minetest.Client, args []string) {
		pos := ao.AOPos{Pos: c.GetPos().Pos.Pos}
		chat.SendMsgf(c, mt.SysMsg, "Got pos: %v", pos)

		_ao := MakeTNT(pos)
		chat.SendMsgf(c, mt.SysMsg, "Made AO: %v", _ao)

		id := ao.RegisterAO(_ao)
		chat.SendMsgf(c, mt.SysMsg, "Got ID: %v", id)
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

func MakeTNT(pos ao.AOPos) *AOTNT {
	return &AOTNT{
		Pos: pos,

		SpawnTime: time.Now(),
	}
}

type AOTNT struct {
	sync.RWMutex

	AOID mt.AOID

	Pos ao.AOPos

	SpawnTime time.Time
}

func (tnt *AOTNT) SetAO(i mt.AOID) {
	tnt.Lock()
	defer tnt.Unlock()

	tnt.AOID = i
}

func (tnt *AOTNT) GetAO() mt.AOID {
	tnt.RLock()
	defer tnt.RUnlock()

	return tnt.AOID
}

func (tnt *AOTNT) SetAOPos(p ao.AOPos) {
	tnt.Lock()
	defer tnt.Unlock()

	tnt.Pos = p
}

func (tnt *AOTNT) GetAOPos() ao.AOPos {
	tnt.RLock()
	defer tnt.RUnlock()

	return tnt.Pos
}

func (tnt *AOTNT) Clean() {
	log.Verbosef("Removing TNT (aoid. %d)\n", tnt.AOID)
}

func (tnt *AOTNT) AOInit(c *minetest.Client) *ao.AOInit {
	tnt.RLock()
	defer tnt.RUnlock()

	return &ao.AOInit{
		AOPos: tnt.Pos,
		HP:    10,

		AOMsgs: []mt.AOMsg{
			&mt.AOCmdProps{
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
			},
			&mt.AOCmdAttach{},
		},
	}
}
