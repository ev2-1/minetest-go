package mapLoader

import (
	"github.com/ev2-1/minetest-go/minetest"

	"time"
)

func lastDim(c *minetest.Client, dim minetest.Dim) *minetest.Dim {
	data, ok := c.GetData("last_dim")
	if !ok {
		c.SetData("last_dim", &dim)

		return nil
	}

	cdim, ok := data.(*minetest.Dim)
	if !ok {
		c.SetData("last_dim", &dim)

		return nil
	}

	c.SetData("last_dim", dim)

	return cdim
}

func init() {
	minetest.RegisterPosUpdater(func(c *minetest.Client, pos *minetest.ClientPos, lu time.Duration) {
		newPos, _ := minetest.Pos2Blkpos(pos.Pos.IntPos())
		oldPos, _ := minetest.Pos2Blkpos(pos.OldPos.IntPos())

		dim := lastDim(c, pos.Pos.Dim)

		if newPos != oldPos || dim == nil || pos.Pos.Dim != *dim {
			c.Logf("Blkpos changed!")

			go loadAround(newPos, c)
		}
	})
}
