package interact

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

func Dig(c *minetest.Client, pos [3]int16) {
	p := minetest.GetPos(c).IntPos()
	p.Pos = pos

	minetest.SetNode(p, mt.Node{Param0: mt.Air}, nil)
}
