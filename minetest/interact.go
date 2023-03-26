package minetest

import (
	"github.com/anon55555/mt"

	"time"
)

func init() {
	RegisterPktProcessor(func(c *Client, pkt *mt.Pkt) {
		i, ok := pkt.Cmd.(*mt.ToSrvInteract)
		if !ok {
			return
		}

		switch pt := i.Pointed.(type) {
		case *mt.PointedNode:
			interactNode(c, i, pt)

		case *mt.PointedAO:
			interactAO(c, i, pt)
		}
	})
}

func interactNode(c *Client, i *mt.ToSrvInteract, pt *mt.PointedNode) {
	if ConfigVerbose() {
		c.Logf("Interact: %s\n", i.Action)
	}

	switch i.Action {
	case mt.Dig: // or hit
		ipos := &IntPos{pt.Under, c.GetPos().Dim}
		c.setDigPos(ipos)

	case mt.StopDigging:
		c.setDigPos(nil)

	case mt.Dug:
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
				Loggers.Warnf("DigPredict '%s' for '%s' is not defined\n", 1, def.Name, def.DigPredict)
				return
			}

			SetNode(IntPos{ptpos, cpos.Dim}, mt.Node{Param0: predict.Thing.Param0, Param1: 255}, nil)
		}

	case mt.Place:
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

func interactAO(c *Client, i *mt.ToSrvInteract, ao *mt.PointedAO) {
	if i.Action != mt.Dig { // unexpected
		c.Logf("Unexpected interaction with AO(%d): %s\n", ao.ID, i.Action)

		return
	}

	_ao := GetAO(ao.ID)

	// get corresponding AO
	// check if client has AO:
	if !HasAO(c, ao.ID) {
		c.Logf("Unexpeted interaction with AO(%d) client does not have AO! (ao is type %T)", ao.ID, _ao)

		return
	}

	if _ao == nil {
		c.Logf("Client either had <nil> AO or interacted with invalid AO")

		return
	}

	_ao.Punch(c, i)
}
