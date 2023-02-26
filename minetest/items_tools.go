package minetest

import (
	"github.com/anon55555/mt"
)

type ToolItem struct {
	Item

	AttackCooldown float32
	MaxDropLvl     int16 //TODO: figure out

	GroupCaps map[string]ToolGroupCap
}

func (itm *ToolItem) Name() string {
	return itm.Item.Name
}

func (itm *ToolItem) ItemDef() mt.ItemDef {
	def := itm.Item.ItemDef()

	def.Type = mt.ToolItem
	def.ToolCaps = mt.ToolCaps{
		NonNil: true,

		AttackCooldown: itm.AttackCooldown,
		MaxDropLvl:     itm.MaxDropLvl,

		GroupCaps: itm.MtGroupCaps(),
	}

	return def
}

func (itm *ToolItem) MtGroupCaps() (s []mt.ToolGroupCap) {
	s = make([]mt.ToolGroupCap, len(itm.GroupCaps))

	var i int
	for name, cap := range itm.GroupCaps {
		s[i] = mt.ToolGroupCap{
			Name: name,

			Uses: cap.Uses,

			MaxLvl: cap.MaxLvl,
			Times:  cap.MtDigTimes(),
		}

		i++
	}

	return
}

func (itm *ToolItem) OnMove(c *Client, s mt.Stack, a *InvAction) mt.Stack            { return s }
func (itm *ToolItem) OnPlace(c *Client, s mt.Stack, a *mt.ToSrvInteract) mt.Stack    { return s }
func (itm *ToolItem) OnUse(c *Client, s mt.Stack, a *mt.ToSrvInteract) mt.Stack      { return s }
func (itm *ToolItem) OnActivate(c *Client, s mt.Stack, a *mt.ToSrvInteract) mt.Stack { return s }

type ToolGroupCap struct {
	Uses int32

	MaxLvl int16

	Times map[int16]float32
}

func (caps ToolGroupCap) MtDigTimes() (s []mt.DigTime) {
	s = make([]mt.DigTime, len(caps.Times))

	var i int
	for r, t := range caps.Times {
		s[i] = mt.DigTime{
			Rating: r,
			Time:   t,
		}

		i++
	}

	return
}
