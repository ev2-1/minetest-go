package minetest

import (
	"github.com/anon55555/mt"

	"log"
	"time"
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
	c.Logf("Interact: %s\n", i.Action)

	switch i.Action {
	case mt.Dig:
		// get pointed node
		pt, ok := i.Pointed.(*mt.PointedNode)
		if !ok {
			c.Logf("[WARN] tried to Dig %T!\n", pt)
			return
		}

		c.setDigPos(&IntPos{pt.Under, GetPos(c).Dim})

	case mt.StopDigging:
		c.setDigPos(nil)

	case mt.Dug:
		pt, ok := i.Pointed.(*mt.PointedNode)
		if !ok {
			c.Logf("[WARN] tried to Dug %T!\n", pt)
			return
		}

		ptpos := pt.Under

		pos, start := c.DigPos()
		c.setDigPos(nil)
		if pos == nil || ptpos != pos.Pos {
			c.Logf("[WARN] clt tired to bamboozle server. (DigPos == nil || DigPos != DugPos)\n")
			return
		}

		dtime := time.Now().Sub(start)

		cpos := GetPos(c)
		if cpos.Dim != pos.Dim {
			c.Logf("[WARN] clt tired to bamboozle server. (DigPos != DugPos (Dimensions dont match))\n")
			return
		}

		blkipos, _ := Pos2Blkpos(cpos.IntPos())
		if !isLoaded(c, blkipos) {
			return
		}

		if !doDigConds(c, i, dtime) {
			node, _ := GetNode(IntPos{ptpos, cpos.Dim})
			c.SendCmd(&mt.ToCltAddNode{
				Pos:      ptpos,
				Node:     node,
				KeepMeta: true,
			})

			return
		}

		// get node digged:
		node, _ := GetNode(IntPos{ptpos, cpos.Dim})
		param0 := node.Param0

		rdef := GetNodeDefID(param0)
		if rdef == nil {
			return
		}

		def := rdef.Thing
		if def.OnDig != nil {
			def.OnDug(c, i, dtime)
		} else {
			//TODO: anticheat

			//dig predict:
			predict := GetNodeDef(rdef.Thing.DigPredict)
			if predict == nil {
				log.Printf("[WARN] DigPredict '%s' for '%s' is not defined\n", def.Name, def.DigPredict)
				return
			}

			SetNode(IntPos{ptpos, cpos.Dim}, mt.Node{Param0: predict.Thing.Param0, Param1: 255}, nil)
		}

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
			DefaultPlace(c, inv, i, def)
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
