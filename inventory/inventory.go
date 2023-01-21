package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"bytes"
	_ "embed"
	"strings"
)

//go:embed inv.fs
var formspec string

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvInvAction:
			action, err := DeserializeInvAction(strings.NewReader(cmd.Action))
			if err != nil {
				c.Logf("ERROR: %s\n", err)
				break
			}

			switch act := action.(type) {
			case *InvActionMove:
				// current_player can all be handled here
				if act.From.Location == "current_player" && act.To.Location == "current_player" {
					Inv, err := GetInv(c)
					if err != nil {
						c.Logf("Error: %s\n", err)
						return
					}

					Inv.Lock()
					defer Inv.Unlock()

					var moveItem string

					// validate source:
					fromInv := Inv.M[act.From.Name]
					if len(fromInv.Stacks) < act.From.Stack || fromInv.Width < act.From.Stack {
						c.Logf("Error: From:_len(stacks) < stack or width < stack")
						return
					}

					moveItem = fromInv.Stacks[act.From.Stack].Name
					c.Logf("Trying to move %d %s", act.Count, moveItem)

					// ensure quantity
					if fromInv.Stacks[act.From.Stack].Count < act.Count {
						c.Log("Error: stack move count > stack count")
						return
					}

					// validate destination
					toInv := Inv.M[act.To.Name]
					if len(toInv.Stacks) < act.To.Stack || toInv.Width < act.To.Stack {
						c.Log("Error: To:_len(stacks) < stack or width < stack")
						return
					}

					// check if slot is empty or same item:
					if !(toInv.Stacks[act.To.Stack].Count == 0 || toInv.Stacks[act.To.Stack].Name == moveItem) {
						c.Log("Error: destination contains other item!")
						return
					}

					// move:
					fromInv.Stacks[act.From.Stack].Count -= act.Count
					toInv.Stacks[act.To.Stack].Name = moveItem
					toInv.Stacks[act.To.Stack].Count += act.Count

					Inv.M[act.To.Name] = toInv
					Inv.M[act.From.Name] = fromInv

					c.Log("Sucessfully did so :)")

					// updating client:
					str, err := Inv.String()
					if err != nil {
						c.Logf("Error: %s", err)
						return
					}

					c.SendCmd(&mt.ToCltInv{
						Inv: str,
					})
				}
			}

			c.Logf("Action: %s\n", action)
		}
	})

	minetest.RegisterRegisterHook(func(c *minetest.Client) {
		GetInv(c) // Pre-Initialize clients Inventory
	})

	minetest.RegisterInitHook(func(c *minetest.Client) {
		c.Logf("Hi!")

		Inv, err := GetInv(c)
		if err != nil {
			c.Logf("Error: %s\n", err)
			return
		}

		Inv.RLock()
		defer Inv.RUnlock()

		// Send client inventory formspec
		c.SendCmd(&mt.ToCltInvFormspec{
			Formspec: formspec,
		})

		str, err := Inv.String()
		if err != nil {
			c.Logf("Error: %s", err)
			return
		}

		c.SendCmd(&mt.ToCltInv{
			Inv: str,
		})
	})
}

func GetInv(c *minetest.Client) (inv *RWInv, err error) {
	data, ok := c.GetData("inv")
	if !ok { // => not found, so initialize
		c.Logf("Client does not have inventory yet, adding")

		stacks := make([]mt.Stack, 4*8)
		stacks[5] = mt.Stack{
			Count: 69,
			Item: mt.Item{
				Name: "basenodes:cobble",
			},
		}

		// Send client inventory contents
		inv = &RWInv{
			Inv: &Inv{
				M: map[string]*mt.InvList{
					"main": &mt.InvList{
						Width:  4 * 8,
						Stacks: stacks,
					},
				},
			},
		}

		c.SetData("inv", inv)

		return inv, nil
	}

	if inv, ok = data.(*RWInv); ok {
		return inv, nil
	}

	if dat, ok := data.(*minetest.ClientDataSaved); ok {
		inv = new(RWInv)

		buf := bytes.NewBuffer(dat.Bytes())

		err = inv.Deserialize(buf)
		c.SetData("inv", inv)

		return inv, err
	}

	return nil, minetest.ErrClientDataInvalidType
}
