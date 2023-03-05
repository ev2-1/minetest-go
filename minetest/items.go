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

// Do not confuse with mt.ItemType
type ItemType uint8

//go:generate stringer --type ItemType
const (
	TypeNodeItem       ItemType = iota
	TypeSimpleNodeItem          //TODO
	TypeCraftItem
	TypeToolItem

	TypeInvalid = 255
)

// Mapps mt.ItemTypes to ItemTypes
func Mt2ItemType(t mt.ItemType) ItemType {
	switch t {
	case mt.NodeItem:
		return TypeSimpleNodeItem
	case mt.CraftItem:
		return TypeCraftItem
	case mt.ToolItem:
		return TypeToolItem
	default:
		return TypeInvalid
	}
}

func (t ItemType) MtItemType() mt.ItemType {
	switch t {
	case TypeNodeItem, TypeSimpleNodeItem:
		return mt.NodeItem
	case TypeCraftItem:
		return mt.CraftItem

	case TypeToolItem:
		return mt.ToolItem
	}

	return 255
}

type Item struct {
	Type ItemType

	Name      string
	Desc      string
	ShortDesc string

	StackMax uint16

	Usable          bool
	CanPointLiquids bool

	Groups map[string]int16

	PointRange float32

	Textures ItemTextures
}

func (itm *Item) MtGroups() (s []mt.Group) {
	s = make([]mt.Group, len(itm.Groups))

	var i int
	for n, r := range itm.Groups {
		s[i] = mt.Group{Name: n, Rating: r}

		i++
	}

	return
}

type ItemDef interface {
	ItemDef() mt.ItemDef
	Name() string

	//returned mt.Stack will overwrite old
	OnMove(*Client, mt.Stack, *InvAction) mt.Stack
	OnPlace(*Client, mt.Stack, *mt.ToSrvInteract) mt.Stack
	OnUse(*Client, mt.Stack, *mt.ToSrvInteract) mt.Stack
	OnActivate(*Client, mt.Stack, *mt.ToSrvInteract) mt.Stack
}

// returns partial mt.ItemDef
func (itm *Item) ItemDef() mt.ItemDef {
	if itm == nil {
		panic("itm == nil")
	}

	return mt.ItemDef{
		Name:      itm.Name,
		Desc:      itm.Desc,
		ShortDesc: itm.ShortDesc,

		InvImg:   TextureStr(itm.Textures.InvImg),
		WieldImg: TextureStr(itm.Textures.WieldImg),

		WieldScale: itm.Textures.WieldScale,

		StackMax: itm.StackMax,

		Usable:          itm.Usable,
		CanPointLiquids: itm.CanPointLiquids,

		Groups: itm.MtGroups(),

		PointRange: itm.PointRange,

		Palette: TextureStr(itm.Textures.Palette),
		Color:   itm.Textures.Color,

		InvOverlay:   TextureStr(itm.Textures.InvOverlay),
		WieldOverlay: TextureStr(itm.Textures.WieldOverlay),
	}
}

var (
	itemNetCacheOnce sync.Once
	itemNetCache     []mt.ItemDef
)

func makeItemNetCache() {
	itemNetCacheOnce.Do(func() {
		itemDefsMu.RLock()
		defer itemDefsMu.RUnlock()

		itemNetCache = make([]mt.ItemDef, len(itemDefs))

		var i int
		for _, rdef := range itemDefs {
			itemNetCache[i] = rdef.Thing.ItemDef()

			i++
		}
	})
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
	itemDefs   = make(map[string]Registerd[ItemDef])
)

// Add more item definitions to pool
func AddItemDef(defs ...ItemDef) {
	itemDefsMu.Lock()
	defer itemDefsMu.Unlock()

	for _, def := range defs {
		// Default debug defs:

		itemDefs[def.Name()] = Registerd[ItemDef]{Caller(1), def}
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
	// add aliases to slice
	aliasesMu.RLock()
	alias := make([]struct{ Alias, Orig string }, len(aliases))

	i := 0
	for k, v := range aliases {
		alias[i] = Alias{k, v}
		i++
	}

	aliasesMu.RUnlock()

	cmd := &mt.ToCltItemDefs{
		Defs:    itemNetCache,
		Aliases: alias,
	}

	return c.SendCmd(cmd)
}

// returns nil if error is encountered
func getItem(c *Client, slot int) (d *Registerd[ItemDef], s mt.Stack, ch chan mt.Stack) {
	ch = make(chan mt.Stack, 1)

	inv, err := GetInv(c)
	if err != nil {
		c.Log("Error during GetInv trying to Place: %s\n", err)
		return
	}

	inv.Lock()
	//TODO: figure out how to do defer inv.Unlock()

	l, ok := inv.Get("main")
	if !ok {
		c.Log("Error: main inv does not exist\n")
		inv.Unlock()
		return
	}

	s, ok = l.GetStack(slot)
	if !ok {
		c.Log("Error: cant get slot %d on main inv\n", slot)
		inv.Unlock()
		return
	}

	if s.Count == 0 {
		c.Log("Error: tried to place 0 stack slot: %d\n", slot)
		inv.Unlock()
		return
	}

	if s.Name == "" {
		c.Log("Error: tried to place stack with no name slot: %d\n", slot)
		inv.Unlock()
		return
	}

	item := GetItemDef(s.Name)
	if item == nil {
		c.Log("Error: tried to place item without definition! name: %s\n", s.Name)
		inv.Unlock()
		return
	}

	go func() {
		defer inv.Unlock()
		defer close(ch)

		stack, ok := <-ch
		if !ok {
			return
		}

		l.SetStack(slot, stack)

		Update(inv, &InvLocation{
			Identifier: &InvIdentifierCurrentPlayer{},
			Name:       "main",
			Stack:      int(slot),
		}, c)
	}()

	return item, s, ch
}

// Trys to create ItemDef from mt.ItemDef
func TryItemDef(def mt.ItemDef) (rdef ItemDef, ok bool) {
	//Basic Information
	item := Item{
		Type: Mt2ItemType(def.Type),

		Name:      def.Name,
		Desc:      def.Desc,
		ShortDesc: def.ShortDesc,

		StackMax: def.StackMax,

		Usable:          def.Usable,
		CanPointLiquids: def.CanPointLiquids,

		Groups: GroupsS2GroupsM(def.Groups),

		PointRange: def.PointRange,

		Textures: ItemTextures{
			Palette: StrTexture(def.Palette),
			Color:   def.Color,

			InvImg:     StrTexture(def.InvImg),
			WieldImg:   StrTexture(def.WieldImg),
			WieldScale: def.WieldScale,
		},
	}

	switch def.Type {
	case mt.NodeItem:
		return &NodeItem{
			Item: item,

			Places:       def.PlacePredict,
			PlaceSnd:     SoundDef(def.PlaceSnd),
			PlaceFailSnd: SoundDef(def.PlaceFailSnd),
		}, true

	case mt.CraftItem:
		return &CraftItem{
			Item: item,
		}, true

	case mt.ToolItem:
		return &ToolItem{
			Item: item,

			AttackCooldown: def.ToolCaps.AttackCooldown,
			MaxDropLvl:     def.ToolCaps.MaxDropLvl,

			GroupCaps: GroupCapsS2ToolGroupCapM(def.ToolCaps.GroupCaps),
		}, true

	default:
		return nil, false
	}
}

func GroupsS2GroupsM(groups []mt.Group) (m map[string]int16) {
	m = make(map[string]int16)
	for _, group := range groups {
		m[group.Name] = group.Rating
	}

	return
}

func GroupCapsS2ToolGroupCapM(s []mt.ToolGroupCap) (m map[string]ToolGroupCap) {
	m = make(map[string]ToolGroupCap)
	for _, cap := range s {
		m[cap.Name] = ToolGroupCap{
			Uses:   cap.Uses,
			MaxLvl: cap.MaxLvl,
			Times:  TimesS2TimesM(cap.Times),
		}
	}

	return
}

func TimesS2TimesM(s []mt.DigTime) (m map[int16]float32) {
	m = make(map[int16]float32)
	for _, time := range s {
		m[time.Rating] = time.Time
	}

	return
}
