package item_ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/minetest/log"

	"image/color"
	"math"
	"strconv"
	"sync"
)

const DropDistance float32 = 12
const a = math.Pi / 180

func getEyeHeight(clt *minetest.Client) float32 {
	data := clt.AOData
	data.RLock()
	defer data.RUnlock()

	return data.SelfProps.EyeHeight
}

func init() {
	minetest.RegisterDropHook(func(clt *minetest.Client, stack mt.Stack, act *minetest.InvActionDrop) mt.Stack {
		apos := clt.GetPos().AOPos()
		pos := apos

		pos.Pos[1] += getEyeHeight(clt) * 10

		pos.Pos[0] += -minetest.Sin32(a*pos.Rot[1]) * DropDistance
		pos.Pos[2] += minetest.Cos32(a*pos.Rot[1]) * DropDistance

		//		pos.Pos[1] += -minetest.Sin32(a*pos.Rot[0]) * DropDistance

		pos.Rot = [3]float32{}

		clt.Logf("Dropped %d %s's %v -> %v\n", act.Count, stack.Name, apos, pos)

		aoid := SpawnItem(pos, stack.Name, int(act.Count))
		if aoid == 0 {
			log.Warnf("Couln't drop '%s' got aoid %d\n", stack.Name, aoid)
		}

		return stack
	})

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

		pos := c.GetPos().AOPos()

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

	Pos minetest.AOPos

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
func (item *ItemAO) SetPos(p minetest.AOPos) {
	item.Lock()
	item.Pos = p
	item.Unlock()

	item.Pos = p

	minetest.BroadcastAOMsgs(item,
		&mt.AOCmdPos{
			Pos: item.Pos.AOPos(),
		},
	)
}

func (item *ItemAO) GetPos() minetest.AOPos {
	item.RLock()
	defer item.RUnlock()

	return item.Pos
}

func SpawnItem(p minetest.AOPos, name string, cnt int) mt.AOID {
	// check name

	p.Pos[1] += 0.15

	return minetest.RegisterAO(&ItemAO{
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
		minetest.RmAO(item.AOID)
	}
}

// compare: minetest-root/builtin/game/item_entity.lua:21
func (item *ItemAO) AOInit(clt *minetest.Client) *minetest.AOInit {
	item.RLock()
	defer item.RUnlock()

	return &minetest.AOInit{
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
