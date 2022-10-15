package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"bufio"
	"bytes"
	"fmt"
)

func init() {
	minetest.RegisterPktProcessor(func(c *minetest.Client, pkt *mt.Pkt) {
		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvInvAction:
			c.Log("Action:", cmd.Action)

			action, err := DeserializeAction(bufio.NewReader(bytes.NewBuffer([]byte(cmd.Action))))
			if err != nil {
				c.Log("Error Parsing action:", err)
				return
			}

			c.Log("action is:", action.Serialize())

			err = action.Apply(c)
			if err != nil {
				c.Log("Error Applying InvAction: " + err.Error())
				return
			}

		case *mt.ToSrvInvFields:
			c.Log(fmt.Sprintf("Form: '%s': '%s'", cmd.Formname, cmd.Fields))
		}
	})
}
