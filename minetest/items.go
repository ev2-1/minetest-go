package minetest

import (
	"github.com/anon55555/mt"

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

func doPlaceConds(clt *Client, i *mt.ToSrvInteract) bool {
	placeCondsMu.RLock()
	defer placeCondsMu.RUnlock()

	for c := range placeConds {
		if !c.Thing(clt, i) {
			return false
		}
	}

	return true
}

type PlaceCond func(*Client, *mt.ToSrvInteract) bool

var (
	placeConds   = make(map[*Registerd[PlaceCond]]struct{})
	placeCondsMu sync.RWMutex
)

// PlaceCond gets called before Place is acted upon
// If returns false doesn't place node (ItemDef.OnPlace not called)
// Gets called BEFORE NodeDef.OnPlace
func RegisterPlaceCond(h PlaceCond) HookRef[Registerd[PlaceCond]] {
	placeCondsMu.Lock()
	defer placeCondsMu.Unlock()

	r := &Registerd[PlaceCond]{Caller(1), h}
	ref := HookRef[Registerd[PlaceCond]]{&placeCondsMu, placeConds, r}

	placeConds[r] = struct{}{}

	return ref
}

var (
	itemDefsMu sync.RWMutex
	itemDefs   = map[string]Registerd[ItemDef]{}
)

// Add more item definitions to pool
func AddItemDef(defs ...ItemDef) {
	itemDefsMu.Lock()
	defer itemDefsMu.Unlock()

	for _, def := range defs {
		// Default debug defs:

		itemDefs[def.Name] = Registerd[ItemDef]{Caller(1), def}
	}
}

// GetItemDef returns pointer to ItemDef if registerd
func GetItemDef(name string) (def *Registerd[ItemDef]) {
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
		defs[i] = v.Thing.ItemDef
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

// returns false if error is encountered
func getItem(c *Client, slot int) (d *Registerd[ItemDef], inv RWInv) {
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

	stack, ok := l.GetStack(slot)
	if !ok {
		c.Logger.Printf("Error: cant get slot %d on main inv\n", slot)
		return
	}

	if stack.Count == 0 {
		c.Logger.Printf("Error: tried to place 0 stack slot: %d\n", slot)
		return
	}

	if stack.Name == "" {
		c.Logger.Printf("Error: tried to place stack with no name slot: %d\n", slot)
		return
	}

	item := GetItemDef(stack.Name)
	if item == nil {
		c.Logger.Printf("Error: tried to place item without definition! name: %s\n", stack.Name)
		return
	}

	return item, inv
}

func DefaultPlace(c *Client, inv RWInv, i *mt.ToSrvInteract, def ItemDef) {
	// Check if item is placable
	if def.PlacePredict == "" {
		return
	}

	if !UseItem(inv, "main", int(i.ItemSlot), 1) {
		return
	}

	Update(inv, &InvLocation{
		Identifier: &InvIdentifierCurrentPlayer{},
		Name:       "main",
		Stack:      int(i.ItemSlot),
	}, c)

	ndef := GetNodeDef(def.PlacePredict)
	if ndef == nil {
		c.Logf("Error in DefaultPlace: PlacePredict of '%s' is not valid node '%s'\n", def.Name, def.PlacePredict)
		return
	}

	param0 := ndef.Thing.Param0

	SetNode(IntPos{i.Pointed.(*mt.PointedNode).Above, c.GetPos().Dim},
		mt.Node{Param0: param0}, nil)
}
