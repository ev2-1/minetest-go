package main

import (
	"github.com/anon55555/mt"
	_minetest "github.com/ev2-1/minetest-go"
	"github.com/ev2-1/minetest-go/abstract"

	"log"
	"time"
)

var posCh <-chan *minetest.CltPos

var (
	MapBlkUpdateRate int64 = 2 // in seconds
	EmptyBlk         mt.MapBlk
)

func init() {
	posCh = minetest.GetPosCh()

	OpenDB(_minetest.Path("/map.sqlite"))

	exampleBlk := mt.MapBlk{}

	for i := 0; i < 4096; i++ {
		exampleBlk.Param0[i] = 126
	}

	for i := 0; i < 16*16; i++ {
		exampleBlk.Param0[i] = 349
	}

	// center block is stone:
	exampleBlk.Param0[4096/2+16/2] = 349 // some wool

	for k := range EmptyBlk.Param0 {
		EmptyBlk.Param0[k] = mt.Air
	}

	/*box := mt.Box{mt.Vec{-0.5, -0.5, -0.5}, mt.Vec{.5, .5, .5}}

	nbox := mt.NodeBox{
		Type: mt.CubeBox,

		WallTop:   box,
		WallBot:   box,
		WallSides: box,
	}

	_minetest.AddNodeDef(mt.NodeDef{
		Param0: 1000,

		Name:     "stone",
		P1Type:   mt.P1Nothing,
		P2Type:   mt.P2Nibble,
		DrawType: mt.DrawCube,

		Scale: 1,

		Tiles: [6]mt.TileDef{
			mt.TileDef{
				Scale: 1,

				Texture: "stone.png",
			},
			mt.TileDef{
				Scale: 1,

				Texture: "stone.png",
			},
			mt.TileDef{
				Scale: 1,

				Texture: "stone.png",
			},
			mt.TileDef{
				Scale: 1,

				Texture: "stone.png",
			},
			mt.TileDef{
				Scale: 1,

				Texture: "stone.png",
			},
			mt.TileDef{
				Scale: 1,

				Texture: "stone.png",
			},
		},

		Color: color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},

		Waving: mt.NotWaving,

		GndContent: true,
		Collides:   true,
		Pointable:  true,
		Diggable:   true,

		LiquidType: mt.NotALiquid,

		DrawBox: nbox,
		ColBox:  nbox,
		SelBox:  nbox,

		Level: 128,
	})*/

	go func() {
		for {
			pos, ok := <-posCh
			if !ok {
				log.Print("[ERROR]", "mapblk pos chan not ok")
				return
			}

			if time.Now().Unix() < pos.LastUpdate+MapBlkUpdateRate {
				p := Pos2int(pos.Pos())
				blkpos, _ := mt.Pos2Blkpos(p)

				blkdata, ok := GetBlk(blkpos)
				if !ok {
					SetBlk(blkpos, &exampleBlk)
					blkdata = exampleBlk
				}

				pos.SendCmd(&mt.ToCltBlkData{
					Blkpos: blkpos,
					Blk:    blkdata,
				})
			}
		}
	}()

	// interactions:
	initInteractions()
}
