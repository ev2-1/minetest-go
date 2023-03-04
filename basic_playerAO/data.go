package playerAO

import (
	"github.com/ev2-1/minetest-go/minetest"

	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"

	"image/color"
	"sync"
)

func init() {
	ao.RegisterPlayerMaker("mcl2", func(clt *minetest.Client, id mt.AOID) ao.ActiveObject {
		return &AOPlayer{
			AOID: id,
			Name: clt.Name,
		}
	})
}

type AOPlayer struct {
	sync.RWMutex

	AOID mt.AOID

	Pos minetest.PPos

	Name string
}

// Implement ao.ActiveObjectRelevant
func (player *AOPlayer) Relevant(clt *minetest.Client) bool {
	cd := ao.GetClientData(clt)

	aopos := player.GetPos()
	cltpos := clt.GetPos()

	cd.RLock()
	defer cd.RUnlock()

	return aopos.Dim == cltpos.Dim &&
		cd.AOID != player.GetAO() &&
		ao.Distance(aopos.Pos.Pos, cltpos.Pos.Pos) <= ao.RelevantDistance
}

func (player *AOPlayer) SetAO(i mt.AOID) {
	player.Lock()
	defer player.Unlock()

	player.AOID = i
}

func (player *AOPlayer) GetAO() mt.AOID {
	player.RLock()
	defer player.RUnlock()

	return player.AOID
}

func (player *AOPlayer) SetPos(p minetest.PPos) {
	player.Lock()
	player.Pos = p
	player.Unlock()

	aopos := ao.PPos2AOPos(p).AOPos()
	aopos.Interpolate = true // if you do a ~360 it still doesn't spin around...

	ao.BroadcastAOMsgs(player,
		&mt.AOCmdPos{
			Pos: aopos,
		},
		&mt.AOCmdBonePos{
			Bone: "Head_Control",
			Pos: mt.AOBonePos{
				Pos: mt.Vec{0, 6.3, 0},
				Rot: mt.Vec{-p.Pitch, 0, 0},
			},
		})
}

func (player *AOPlayer) GetPos() minetest.PPos {
	player.RLock()
	defer player.RUnlock()

	return player.Pos
}

func (player *AOPlayer) Clean() {}

func (player *AOPlayer) AOInit(clt *minetest.Client) *ao.AOInit {
	player.RLock()
	defer player.RUnlock()

	return &ao.AOInit{
		Name:     player.Name,
		IsPlayer: true,

		AOPos: ao.PPos2AOPos(player.Pos),
		HP:    10,

		AOMsgs: []mt.AOMsg{
			&mt.AOCmdBonePos{
				Bone: "Body_Control", Pos: mt.AOBonePos{
					Pos: mt.Vec{0, 6.3, 0},
					Rot: mt.Vec{0, 0, 0},
				},
			},
			&mt.AOCmdBonePos{
				Bone: "Head_Control", Pos: mt.AOBonePos{
					Pos: mt.Vec{0, 6.3, 0},
					Rot: mt.Vec{0, 0, 0},
				},
			},
			&mt.AOCmdBonePos{
				Bone: "Arm_Right_Pitch_Control", Pos: mt.AOBonePos{
					Pos: mt.Vec{-3, 5.785, 0},
					Rot: mt.Vec{0, 0, 0},
				},
			},
			&mt.AOCmdBonePos{
				Bone: "Arm_Left_Pitch_Control", Pos: mt.AOBonePos{
					Pos: mt.Vec{3, 5.785, 0},
					Rot: mt.Vec{0, 0, 0},
				},
			},
			&mt.AOCmdBonePos{
				Bone: "Wield_Item", Pos: mt.AOBonePos{
					Pos: mt.Vec{-1.5, 4.9, 1.8},
					Rot: mt.Vec{135, 0, 90},
				},
			},
			&mt.AOCmdPhysOverride{
				Phys: mt.AOPhysOverride{
					Walk:    1,
					Jump:    1,
					Gravity: 1,
				},
			},
			&mt.AOCmdProps{
				Props: mt.AOProps{
					MaxHP:      20,
					ColBox:     mt.Box{mt.Vec{-0.312, 0, -0.312}, mt.Vec{0.312, 1.8, 0.312}},
					SelBox:     mt.Box{mt.Vec{-0.312, 0, -0.312}, mt.Vec{0.312, 1.8, 0.312}},
					Pointable:  true,
					Visual:     "mesh",
					VisualSize: mt.Vec{1, 1, 1},

					Visible:  true,
					Textures: []mt.Texture{"mcl_skins_character_1.png", "blank.png", "blank.png"},

					SpriteSheetSize:  [2]int16{1, 1},
					SpritePos:        [2]int16{0, 0},
					MakeFootstepSnds: true,
					RotateSpeed:      0,
					Mesh:             "mcl_armor_character_female.b3d",

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
			&mt.AOCmdAttach{
				Attach: mt.AOAttach{
					ParentID: 0,
				},
			},
		},
	}
}
