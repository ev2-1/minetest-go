package inventory

import (
	//	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

var updateInv = make(chan Inv, 16)

func init() {
	minetest.RegisterStage2(func() {
		// TODO: do the updating
	})
}
