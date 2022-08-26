package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
)

func init() {
	minetest.RegisterPktProcessor(func(clt *minetest.Client, pkt *mt.Pkt) {
		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvChatMsg:
			switch cmd.Msg {
			case "list_aos":
				clt.SendCmd(&mt.ToCltChatMsg{
					Type: mt.RawMsg,
					Text: fmt.Sprintf("aos: %v", activeObjects),
				})
				break

			case "list_my_aos":
				clientsMu.RLock()

				clt.SendCmd(&mt.ToCltChatMsg{
					Type: mt.RawMsg,
					Text: fmt.Sprintf("Your (%s) aos: %v", clt.Name, clients[clt].aos),
				})

				clientsMu.RUnlock()
				break
			}
			break
		}
	})
}
