package item_ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"

	"image/color"
	"math"
	"strconv"
	"sync"
)

func init() {
	chat.RegisterChatCmd("spawn_item", func(c *minetest.Client, args []string) {
		usage := func() {
			chat.SendMsgf(c, mt.NormalMsg, "Usage: spawn_item <item> [count]")
		}

		if len(args) < 2 {
			usage()
			return
		}

		count, err := strconv.Atoi(args[1])
		if err != nil {
			usage()
			return
		}

		if count <= 0 {
			usage()
			return
		}

		pos := ao.PPos2AOPos(c.GetPos())

		aoid := SpawnItem(pos, args[0], count)
		if aoid == 0 {
			chat.SendMsgf(c, mt.NormalMsg, "Could not spawn item with name '%s'", args[0])
		} else {
			chat.SendMsgf(c, mt.NormalMsg, "Spawn item with name '%s'; got AOID %d", args[0], aoid)
		}
	})
}

type ItemAO struct {
	sync.RWMutex

	AOID mt.AOID

	Pos ao.AOPos

	Count int
	Name  string
}

func (item *ItemAO) Clean() {
}

func (item *ItemAO) SetAO(i mt.AOID) {
	item.Lock()
	defer item.Unlock()

	item.AOID = i
}

func (item *ItemAO) GetAO() mt.AOID {
	item.RLock()
	defer item.RUnlock()

	return item.AOID
}
func (item *ItemAO) SetPos(p ao.AOPos) {
	item.Lock()
	item.Pos = p
	item.Unlock()

	item.Pos = p

	ao.BroadcastAOMsgs(item,
		&mt.AOCmdPos{
			Pos: item.Pos.AOPos(),
		},
	)
}

func (item *ItemAO) GetPos() ao.AOPos {
	item.RLock()
	defer item.RUnlock()

	return item.Pos
}

func SpawnItem(p ao.AOPos, name string, cnt int) mt.AOID {
	// check name

	p.Pos[1] += 0.15

	return ao.RegisterAO(&ItemAO{
		Pos: p,

		Count: cnt,
		Name:  name,
	})
}

func (item *ItemAO) Punch(clt *minetest.Client, i *mt.ToSrvInteract) {
	if item.Count == 0 || item.Name == "" {
		return
	}

	added, _, err := minetest.Give(clt, &minetest.InvLocation{
		Identifier: &minetest.InvIdentifierCurrentPlayer{},
		Name:       "main",
		Stack:      -1,
	}, uint16(item.Count), item.Name)
	if err != nil {
		clt.Logf("Error giving %d item %s: %s\n", item.Count, item.Name, err)
	}

	item.Count -= int(added)
	clt.Logf("Added %d, new cnt: %d\n", added, item.Count)
	if item.Count == 0 {
		ao.RmAO(item.AOID)
	}
}

// compare: minetest-root/builtin/game/item_entity.lua:21
func (item *ItemAO) AOInit(clt *minetest.Client) *ao.AOInit {
	item.RLock()
	defer item.RUnlock()

	return &ao.AOInit{
		IsPlayer: false,

		AOPos: item.Pos,
		HP:    1,

		AOMsgs: []mt.AOMsg{
			&mt.AOCmdProps{
				Props: mt.AOProps{
					MaxHP:            1,
					CollideWithNodes: true,

					//					ColBox: mt.Box{mt.Vec{-0.3, -0.3, -0.3} mt.Vec{0.3, 0.3, 0.3}},
					SelBox:     mt.Box{mt.Vec{-0.15, -0.15, -0.15}, mt.Vec{0.15, 0.15, 0.15}},
					Pointable:  true,
					Visual:     "wielditem",
					VisualSize: mt.Vec{0.1, 0.1, 0.1},

					Visible:  true,
					Textures: []mt.Texture{mt.Texture(item.Name)}, //TODO lookup?

					SpriteSheetSize:  [2]int16{1, 1},
					SpritePos:        [2]int16{0, 0},
					MakeFootstepSnds: true,
					RotateSpeed:      math.Pi * 0.5 * 0.2,

					Colors: []color.NRGBA{color.NRGBA{R: 255, G: 255, B: 255, A: 255}},

					CollideWithAOs: true,
					StepHeight:     6,
					NametagColor:   color.NRGBA{R: 255, G: 255, B: 255, A: 255},

					FaceRotateSpeed: -1,
					MaxBreath:       10,
					EyeHeight:       1.5,
					Shaded:          true,
					ShowOnMinimap:   true,
				},
			},
			&mt.AOCmdAttach{},
		},
	}
}
