package playerAO

import (
	"github.com/ev2-1/minetest-go/minetest"

	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"

	"image/color"
)

type player struct {
	ao.ActiveObjectS

	name string
}

func (p *player) InitPkt(id mt.AOID, clt *minetest.Client) mt.AOInitData {
	data := p.ActiveObjectS.InitPkt(id, clt)

	data.Name = p.name
	data.IsPlayer = true

	data.Pos = mt.Pos{0, 100, 0}

	return data
}

func makeAO(clt *minetest.Client) ao.ActiveObject {
	return &player{
		name: clt.Name,

		ActiveObjectS: ao.ActiveObjectS{
			AOState: ao.AOState{
				Bones: map[string]mt.AOBonePos{
					"Body_Control": mt.AOBonePos{
						Pos: mt.Vec{0, 6.3, 0},
						Rot: mt.Vec{0, 0, 0},
					},
					"Head_Control": mt.AOBonePos{
						Pos: mt.Vec{0, 6.3, 0},
						Rot: mt.Vec{0, 0, 0},
					},
					"Arm_Right_Pitch_Control": mt.AOBonePos{
						Pos: mt.Vec{-3, 5.785, 0},
						Rot: mt.Vec{0, 0, 0},
					},
					"Arm_Left_Pitch_Control": mt.AOBonePos{
						Pos: mt.Vec{3, 5.785, 0},
						Rot: mt.Vec{0, 0, 0},
					},
					"Wield_Item": mt.AOBonePos{
						Pos: mt.Vec{-1.5, 4.9, 1.8},
						Rot: mt.Vec{135, 0, 90},
					},
				},

				Phys: ao.AOPhys{
					Walk:    1,
					Jump:    1,
					Gravity: 1,
				},
			},

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
	}
}
