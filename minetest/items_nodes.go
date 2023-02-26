package minetest

import (
	"github.com/anon55555/mt"
)

// NodeItem specifies a item that will palce NoteItem.Places
type NodeItem struct {
	Item

	Places string

	PlaceSnd, PlaceFailSnd SoundDef
}

func (itm *NodeItem) Name() string { return itm.Item.Name }

func (itm *NodeItem) ItemDef() mt.ItemDef {
	def := itm.Item.ItemDef()

	def.Type = mt.NodeItem
	def.PlacePredict = itm.Places
	def.PlaceSnd = mt.SoundDef(itm.PlaceSnd)
	def.PlaceFailSnd = mt.SoundDef(itm.PlaceFailSnd)

	return def
}

func (itm *NodeItem) OnMove(c *Client, s mt.Stack, a *InvAction) mt.Stack { return s }
func (itm *NodeItem) OnPlace(c *Client, s mt.Stack, i *mt.ToSrvInteract) mt.Stack {
	// Check if item is placable
	if itm.Places == "" {
		return s
	}

	ndef := GetNodeDef(itm.Places)
	if ndef == nil {
		c.Logf("PlacePredict of item '%s' is not registered (PlacePredict is '%s')\n",
			itm.Name(), itm.Places,
		)

		return s
	}

	param0 := ndef.Thing.Param0

	if s.Count <= 0 {
		return s
	}

	s.Count--

	SetNode(IntPos{i.Pointed.(*mt.PointedNode).Above, c.GetPos().Dim},
		mt.Node{Param0: param0}, nil)

	return s
}
func (itm *NodeItem) OnUse(c *Client, s mt.Stack, a *mt.ToSrvInteract) mt.Stack      { return s }
func (itm *NodeItem) OnActivate(c *Client, s mt.Stack, a *mt.ToSrvInteract) mt.Stack { return s }

var _ ItemDef = &NodeItem{}
