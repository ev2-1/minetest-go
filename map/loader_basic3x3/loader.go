package mapLoader

import (
	"github.com/ev2-1/minetest-go/minetest"

	"time"
)

func lastDim(c *minetest.Client, dim minetest.DimID) *minetest.DimID {
	data, ok := c.GetData("last_dim")
	if !ok {
		c.SetData("last_dim", dim)

		return nil
	}

	cdim, ok := data.(minetest.DimID)
	if !ok {
		c.Logf("last_dim has type %T\n", data)
		c.SetData("last_dim", dim)

		return nil
	}

	c.SetData("last_dim", dim)

	return &cdim
}

func init() {
	minetest.RegisterPosUpdater(func(c *minetest.Client, pos *minetest.ClientPos, lu time.Duration) {
		newPos, _ := minetest.Pos2Blkpos(pos.CurPos.IntPos())
		oldPos, _ := minetest.Pos2Blkpos(pos.OldPos.IntPos())

		dim := lastDim(c, pos.CurPos.Dim)

		if newPos != oldPos || dim == nil || pos.CurPos.Dim != *dim {
			c.Logf("Blkpos changed. %v %v %v\n", newPos != oldPos, newPos != oldPos || dim == nil, newPos != oldPos || dim == nil || pos.CurPos.Dim != *dim)

			go loadAround(newPos, c)
		}
	})
}
