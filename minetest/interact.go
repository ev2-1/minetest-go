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
	if ConfigVerbose() {
		c.Logf("Interact: %s\n", i.Action)
	}

	switch i.Action {
	case mt.Dig: // or hit
		// get pointed node
		pt, ok := i.Pointed.(*mt.PointedNode)
		if !ok {
			return
		}

		ipos := &IntPos{pt.Under, c.GetPos().Dim}
		c.setDigPos(ipos)

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

		cpos := c.GetPos()
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
		if def.OnDug != nil {
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

		dim := c.GetPos().Dim
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
		rdef, stack, uch := getItem(c, int(i.ItemSlot))

		// update
		defer func() { uch <- stack }()

		if rdef == nil {
			return
		}
		def := rdef.Thing

		stack = def.OnPlace(c, stack, i)

	case mt.Use:
		// get item in hand:
		def, stack, uch := getItem(c, int(i.ItemSlot))

		// update
		defer func() { uch <- stack }()
		if def == nil {
			return
		}

		stack = def.Thing.OnUse(c, stack, i)

	case mt.Activate:
		// get item in hand:
		def, stack, uch := getItem(c, int(i.ItemSlot))

		// update
		defer func() { uch <- stack }()
		if def == nil {
			return
		}

		def.Thing.OnActivate(c, stack, i)
	}
}
