package minetest

import (
	"github.com/anon55555/mt"
)

func init() {
	RegisterPktProcessor(func(c *Client, pkt *mt.Pkt) {
		m, ok := pkt.Cmd.(*mt.ToSrvInteract)

		if ok {
			interact(c, m)
		}
	})
}

func interact(c *Client, i *mt.ToSrvInteract) {
	switch i.Action {
	case mt.Place:
		// get pointed node
		pt, ok := i.Pointed.(*mt.PointedNode)
		if !ok {
			c.Logf("[WARN] tried to Place: on %T!\n", pt)
			return
		}

		dim := GetPos(c).Dim
		pos := pt.Above
		blkpos, _ := mt.Pos2Blkpos(pos)
		blkipos := IntPos{blkpos, dim}
		ipos := IntPos{pos, dim}

		if !doPlaceConds(c, i) {
			if isLoaded(c, blkipos) {
				node, _ := GetNode(ipos)

				c.SendCmd(&mt.ToCltAddNode{
					Pos:      pos,
					Node:     node,
					KeepMeta: true,
				})
			}
		}

		//place conditions
		if !doPlaceConds(c, i) {
			return
		}

		// get item in hand:
		rdef, inv := getItem(c, int(i.ItemSlot))
		if rdef == nil {
			return
		}
		def := rdef.Thing

		if def.OnPlace != nil {
			def.OnPlace(c, inv, i)
		} else {

		}

	case mt.Use:
		def, inv := getItem(c, int(i.ItemSlot))
		if def == nil {
			return
		}

		if def.Thing.OnUse != nil {
			def.Thing.OnUse(c, inv, i)
		}
	case mt.Activate:
		def, inv := getItem(c, int(i.ItemSlot))
		if def == nil {
			return
		}

		if def.Thing.OnActivate != nil {
			def.Thing.OnActivate(c, inv, i)
		}
	}
}
