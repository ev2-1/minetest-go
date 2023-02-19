package minetest

import (
	"github.com/anon55555/mt"

	"log"
	"sync"
)

type (
	ItemPlaceFunc    func(*Client, Inv, *mt.ToSrvInteract)
	ItemUseFunc      func(*Client, Inv, *mt.ToSrvInteract)
	ItemActivateFunc func(*Client, Inv, *mt.ToSrvInteract)
	ItemMoveFunc     func(*Client, Inv, *InvAction)
)

type ItemDef struct {
	mt.ItemDef

	OnPlace    ItemPlaceFunc
	OnUse      ItemUseFunc
	OnActivate ItemActivateFunc
	OnMove     ItemMoveFunc
}

var (
	itemDefsMu sync.RWMutex
	itemDefs   = map[string]ItemDef{}
)

func DebugMove(def ItemDef) ItemMoveFunc {
	return func(c *Client, inv Inv, i *InvAction) {}
}

func DebugActivate(def ItemDef) ItemActivateFunc {
	return func(c *Client, inv Inv, i *mt.ToSrvInteract) {}
}

func DebugUse(def ItemDef) ItemUseFunc {
	return func(c *Client, inv Inv, i *mt.ToSrvInteract) {}
}

func DebugPlace(def ItemDef) ItemPlaceFunc {
	return func(c *Client, inv Inv, i *mt.ToSrvInteract) {
		n, ok := i.Pointed.(*mt.PointedNode)
		if !ok {
			return
		}

		pos := n.Above
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

		node := GetNodeDef(def.PlacePredict)
		if node == nil {
			log.Fatalf("Error: item place prediction for %s (%s) does not exists as node\n", stack.Name, def.PlacePredict)
		}

		c.Logger.Printf("Placing %s (param0: %d) at (%d,%d,%d)", def.PlacePredict, node.Param0, pos[0], pos[1], pos[2])

		// Get DIM:
		p := GetPos(c).IntPos()
		p.Pos = pos

		SetNode(p, mt.Node{Param0: node.Param0}, nil)

		// remove item from inv
		stack.Count--
		l.SetStack(int(i.ItemSlot), stack)

		str, err := SerializeString(inv.Serialize)
		if err != nil {
			c.Logger.Printf("Error: cant serialize inv: %s\n", err)
			return
		}

		// ahh yes, i *love* object orientation
		(&InvLocation{
			Identifier: &InvIdentifierCurrentPlayer{},
			Name:       "main",
			Stack:      int(i.ItemSlot),
		}).SendUpdate(str, c)
	}
}

// Add more item definitions to pool
func AddItemDef(defs ...ItemDef) {
	itemDefsMu.Lock()
	defer itemDefsMu.Unlock()

	for _, def := range defs {
		// Default debug defs:
		if def.OnPlace == nil {
			log.Printf("[WARN] ItemDef for %s has to OnPlace, using DebugPlace\n", def.Name)
			def.OnPlace = DebugPlace(def)
		}

		if def.OnUse == nil {
			def.OnUse = DebugUse(def)
		}

		if def.OnActivate == nil {
			def.OnActivate = DebugActivate(def)
		}

		if def.OnMove == nil {
			def.OnMove = DebugMove(def)
		}

		itemDefs[def.Name] = def
	}
}

// GetItemDef returns pointer to ItemDef if registerd
func GetItemDef(name string) (def *ItemDef) {
	itemDefsMu.Lock()
	defer itemDefsMu.Unlock()

	d, found := itemDefs[name]
	if !found {
		return nil
	}

	return &d
}

// Send (cached) ItemDefinitions to client
func (c *Client) SendItemDefs() (<-chan struct{}, error) {
	itemDefsMu.RLock()

	// add to slice
	defs := make([]mt.ItemDef, len(itemDefs))
	var i int
	for _, v := range itemDefs {
		defs[i] = v.ItemDef
		i++
	}

	itemDefsMu.RUnlock()

	// add aliases to slice
	aliasesMu.RLock()
	alias := make([]struct{ Alias, Orig string }, len(aliases))

	i = 0
	for k, v := range aliases {
		alias[i] = Alias{k, v}
		i++
	}

	aliasesMu.RUnlock()

	cmd := &mt.ToCltItemDefs{
		Defs:    defs,
		Aliases: alias,
	}

	return c.SendCmd(cmd)
}
