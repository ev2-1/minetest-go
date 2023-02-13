package gen_flat

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"log"
	"time"
)

func init() {
	minetest.RegisterMapGenerator("flat", &FlatMapGenerator{})
}

type FlatMapGenerator struct {
	Config [][16]string

	Driver minetest.MapDriver
}

func (gen *FlatMapGenerator) FromS(s []string) {
	gen.Config = make([][16]string, 1)

	for k := range s {
		i := k / 16
		if len(gen.Config) <= i {
			gen.Config = append(gen.Config, [16]string{})
		}

		log.Printf("i: %2d; k: %2d;\n", i, k)

		gen.Config[i][k-i*16] = s[k]
	}

	log.Printf("-%v-\n", gen.Config)
}

func (*FlatMapGenerator) Make(drv minetest.MapDriver) minetest.MapGenerator {
	gen := new(FlatMapGenerator)
	gen.Driver = drv

	gen.FromS([]string{
		"mcl_core:bedrock", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt", "mcl_core:dirt_with_grass",
	})

	return gen
}

const length = 16 * 16

func (g *FlatMapGenerator) Generate(pos [3]int16) (*minetest.MapBlk, error) {
	var blk mt.MapBlk
	line := func(z int, c mt.Content) {
		for x := 0; x < 16; x++ {
			for y := 0; y < 16; y++ {
				blk.Param0[x+(16*z)+(y*16*16)] = c
			}
		}
	}

	if pos[1] >= 0 && int(pos[1]) < len(g.Config) {
		s := g.Config[pos[1]]

		for x, name := range s {
			var id mt.Content
			if name == "" {
				id = mt.Air
			} else {
				id = minetest.GetNodeID(name)
			}

			line(x, id)
		}
	} else {
		for k := range blk.Param0 {
			blk.Param0[k] = mt.Air
		}
	}

	mapblk := &minetest.MapBlk{
		MapBlk: blk,
		Pos:    pos,

		Driver: g.Driver,
		Loaded: time.Now(),
	}

	err := mapblk.Save()
	if err != nil {
		return nil, err
	}

	return mapblk, nil
}
