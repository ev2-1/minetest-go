package minetest

import (
	"github.com/anon55555/mt"
)

type CraftItem struct {
	Item
}

func (itm *CraftItem) Name() string { return itm.Item.Name }

func (itm *CraftItem) ItemDef() mt.ItemDef {
	def := itm.Item.ItemDef()
	def.Type = mt.CraftItem

	return def
}

func (itm *CraftItem) OnMove(c *Client, s mt.Stack, a *InvAction) mt.Stack            { return s }
func (itm *CraftItem) OnPlace(c *Client, s mt.Stack, a *mt.ToSrvInteract) mt.Stack    { return s }
func (itm *CraftItem) OnUse(c *Client, s mt.Stack, a *mt.ToSrvInteract) mt.Stack      { return s }
func (itm *CraftItem) OnActivate(c *Client, s mt.Stack, a *mt.ToSrvInteract) mt.Stack { return s }

var _ ItemDef = &CraftItem{}
