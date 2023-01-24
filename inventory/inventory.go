package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"

	"bytes"
	_ "embed"
	"strings"
)

//go:embed inv.fs
var formspec string

func init() {
	RegisterDetached("test", &DetachedInv{
		SimpleInv: SimpleInv{
			M: map[string]InvList{
				"main": &SimpleInvList{
					mt.InvList{
						Width:  4 * 8,
						Stacks: make([]mt.Stack, 4*8),
					},
				},
			},
		},
	})

	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvInvAction:
			action, err := DeserializeInvAction(strings.NewReader(cmd.Action))
			if err != nil {
				c.Logger.Printf("Error: %s", err)
				break
			}

			if _, err := action.Apply(c); err != nil {
				c.Logger.Printf("Error: %s", err)
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
			c.Logger.Printf("Error: %s", err)
			return
		}

		Inv.RLock()
		defer Inv.RUnlock()

		str, err := SerializeString(Inv.Serialize)
		if err != nil {
			c.Logger.Printf("Error: %s", err)
			return
		}

		ack, _ := c.SendCmd(&mt.ToCltInv{
			Inv: str,
		})

		<-ack
		c.Logger.Printf("Sent CltInv")
	})

	chat.RegisterChatCmd("showspec", func(c *minetest.Client, args []string) {
		c.SendCmd(&mt.ToCltShowFormspec{
			Formspec: formspec,
			Formname: "lol",
		})
	})

	chat.RegisterChatCmd("getdetached", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: getdetached [name]", mt.RawMsg)
			return
		}

		d, err := GetDetached(args[0], c)
		if err != nil {
			c.Logger.Printf("Error: %s", err)
			return
		}

		ack, err := d.AddClient(c)
		if err != nil {
			c.Logger.Printf("Error: %s", err)
			return
		}

		<-ack
		c.Logger.Printf("Sent DetachedInv")

	})
}

func GetInv(c *minetest.Client) (inv *SimpleInv, err error) {
	data, ok := c.GetData("inv")
	if !ok { // => not found, so initialize
		c.Logger.Printf("Client does not have inventory yet, adding")

		stacks := make([]mt.Stack, 4*8)
		stacks[5] = mt.Stack{
			Count: 69,
			Item: mt.Item{
				Name: "basenodes:cobble",
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

	if dat, ok := data.(*minetest.ClientDataSaved); ok {
		inv = new(SimpleInv)

		buf := bytes.NewBuffer(dat.Bytes())

		err = inv.Deserialize(buf)
		c.SetData("inv", inv)

		return inv, err
	}

	return nil, minetest.ErrClientDataInvalidType
}
