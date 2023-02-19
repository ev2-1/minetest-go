package interact

import (
	"log"

	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

func Place(c *minetest.Client, i *mt.ToSrvInteract) {
	n, ok := i.Pointed.(*mt.PointedNode)
	if !ok {
		return
	}

	pos := n.Above

	// get item in hand:
	inv, err := minetest.GetInv(c)
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

	item := minetest.GetItemDef(stack.Name)
	if item == nil {
		c.Logger.Printf("Error: tried to place item without definition! name: %s\n", stack.Name)
		return
	}

	node := minetest.GetNodeDef(item.PlacePredict)
	if node == nil {
		log.Fatalf("Error: item place prediction for %s (%s) does not exists as node\n", stack.Name, item.PlacePredict)
	}

	c.Logger.Printf("Placing %s (param0: %d) at (%d,%d,%d)", item.PlacePredict, node.Param0, pos[0], pos[1], pos[2])

	p := minetest.GetPos(c).IntPos()
	p.Pos = pos

	minetest.SetNode(p, mt.Node{Param0: node.Param0}, nil)

	// remove item from inv
	stack.Count--
	l.SetStack(int(i.ItemSlot), stack)

	str, err := minetest.SerializeString(inv.Serialize)
	if err != nil {
		c.Logger.Printf("Error: cant serialize inv: %s\n", err)
		return
	}

	// ahh yes, i *love* object orientation
	(&minetest.InvLocation{
		Identifier: &minetest.InvIdentifierCurrentPlayer{},
		Name:       "main",
		Stack:      int(i.ItemSlot),
	}).SendUpdate(str, c)
}
