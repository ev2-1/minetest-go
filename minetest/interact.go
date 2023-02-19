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
	case mt.Place, mt.Use, mt.Activate:
		// get item in hand:
		inv, err := GetInv(c)
		if err != nil {
			c.Logger.Printf("Error during GetInv trying to Place: %s\n", err)
			return
		}

		inv.Lock()
		defer inv.Unlock()

		l, ok := inv.Get("main")
		if !ok {
			c.Logger.Printf("Error: main inv does not exist\n")
			return
		}

		stack, ok := l.GetStack(int(i.ItemSlot))
		if !ok {
			c.Logger.Printf("Error: cant get slot %d on main inv\n", i.ItemSlot)
			return
		}

		if stack.Count == 0 {
			c.Logger.Printf("Error: tried to place 0 stack slot: %d\n", i.ItemSlot)
			return
		}

		if stack.Name == "" {
			c.Logger.Printf("Error: tried to place stack with no name slot: %d\n", i.ItemSlot)
			return
		}

		item := GetItemDef(stack.Name)
		if item == nil {
			c.Logger.Printf("Error: tried to place item without definition! name: %s\n", stack.Name)
			return
		}

		switch i.Action {
		case mt.Place:
			item.OnPlace(c, inv, i)
		case mt.Use:
			item.OnUse(c, inv, i)
		case mt.Activate:
			item.OnActivate(c, inv, i)

		default:
			panic("How did we get here?")
		}
	}
}
