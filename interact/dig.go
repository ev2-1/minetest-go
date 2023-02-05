package interact

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

func Dig(pos [3]int16) {
	minetest.SetNode(pos, mt.Node{Param0: mt.Air}, nil)
}
