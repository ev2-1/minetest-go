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
				c.Logf("Error: %s", err)
				break
			}

			if _, err := action.Apply(c); err != nil {
				c.Logf("Error: %s", err)
			}
		}
	})

	minetest.RegisterInitHook(func(c *minetest.Client) {
		// Send client inventory formspec
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
