package example

import (
	inv "github.com/ev2-1/minetest-go/inventory"
	"github.com/ev2-1/minetest-go/minetest"

	"github.com/anon55555/mt"

	_ "embed"
)

//go:embed form.spec
var formspec string

type PlayerInv inv.StorageInv

func (i *PlayerInv) Type() inv.InvLocationType {
	return inv.InvLocationCurrentPlayer
}

func init() {
	inv.RegisterPlayerInventoryType("basic", func(c *minetest.Client) inv.PlayerInv {
		stacks := make([]mt.Stack, 32)

		stacks[1] = mt.Stack{Item: mt.Item{Name: "basenodes:dirt_with_grass"}, Count: 69}
		stacks[10] = mt.Stack{Item: mt.Item{Name: "basenodes:dirt_with_grass"}, Count: 420}

		return inv.PlayerInv{
			Inv: map[string]inv.Inv{
				"main": &PlayerInv{
					InvList: mt.InvList{
						Width:  32,
						Stacks: stacks,
					},
				},
				"craft": &inv.StorageInv{
					InvList: mt.InvList{
						Width:  9,
						Stacks: make([]mt.Stack, 9),
					},
				},
				"craftpreview": &inv.StorageInv{
					InvList: mt.InvList{
						Width:  1,
						Stacks: make([]mt.Stack, 1),
					},
				},
			},
			Formspec: formspec,
		}
	})

	inv.SetDefaultInventory("basic")
}
