package minetest

import (
	"github.com/anon55555/mt"

	"bytes"
	_ "embed"
	"strings"
)

// TODO: clean
//
//go:embed inv.fs
var formspec string

func TestSpec() string {
	return formspec
}

func init() {
	RegisterPktProcessor(func(c *Client, pkt *mt.Pkt) {
		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvInvAction:
			action, err := DeserializeInvAction(strings.NewReader(cmd.Action))
			if err != nil {
				c.Logf("Error: %s", err)
				break
			}

			if _, err := action.Apply(c); err != nil {
				c.Logf("Error: %s", err)
			}
		}
	})

	RegisterInitHook(func(c *Client) {
		// Send client inventory formspec
		// TODO: formspecs based on setting in ClientData & config field
		c.SendCmd(&mt.ToCltInvFormspec{
			Formspec: formspec,
		})

		Inv, err := GetInv(c)
		if err != nil {
			c.Logf("Error: %s", err)
			return
		}

		Inv.RLock()
		defer Inv.RUnlock()

		str, err := SerializeString(Inv.Serialize)
		if err != nil {
			c.Logf("Error: %s", err)
			return
		}

		ack, _ := c.SendCmd(&mt.ToCltInv{
			Inv: str,
		})

		<-ack
		c.Log("Sent CltInv")
	})
}

func GetInv(c *Client) (inv *SimpleInv, err error) {
	data, ok := c.GetData("inv")
	if !ok { // => not found, so initialize
		c.Log("Client does not have inventory yet, adding")

		//TODO: clean
		stacks := make([]mt.Stack, 4*8)
		stacks[5] = mt.Stack{
			Count: 69,
			Item: mt.Item{
				Name: "mcl_core:stone",
			},
		}

		// Send client inventory contents
		inv = &SimpleInv{
			M: map[string]InvList{
				"main": &SimpleInvList{
					mt.InvList{
						Width:  4 * 8,
						Stacks: stacks,
					},
				},
			},
		}

		c.SetData("inv", inv)

		return inv, nil
	}

	if inv, ok = data.(*SimpleInv); ok {
		return inv, nil
	}

	if dat, ok := data.(*ClientDataSaved); ok {
		inv = new(SimpleInv)

		buf := bytes.NewBuffer(dat.Bytes())

		err = inv.Deserialize(buf)
		c.SetData("inv", inv)

		return inv, err
	}

	return nil, ErrClientDataInvalidType
}

func UseItem(inv RWInv, name string, slot int, i int) bool {
	inv.Lock()
	defer inv.Unlock()

	list, ok := inv.Get(name)
	if !ok {
		return false
	}

	stack, ok := list.GetStack(slot)
	if !ok {
		return false
	}

	if stack.Count <= 0 {
		return false
	}

	stack.Count -= 1

	ok = list.SetStack(slot, stack)

	return ok
}

func Update(inv RWInv, loc *InvLocation, c *Client) (<-chan struct{}, error) {
	str, err := SerializeString(inv.Serialize)
	if err != nil {
		return nil, err
	}

	return loc.SendUpdate(str, c)
}
