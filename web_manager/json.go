package web_manager

import (
	"encoding/json"
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
)

type Message interface{}

type MsgPacket struct {
	Packet Packet
}

type Packet struct {
	Type string
	Name string
	Clt  string // ip of client
	Srv  bool   // send by server
	Cmd  mt.Cmd
}

func (mp *MsgPacket) MarshalJSON() (data []byte, err error) {
	content, err := json.Marshal(mp.Packet)
	if err != nil {
		return
	}

	return []byte("{Type:\"packet\",Packet:" + string(content) + "}"), nil
}

func init() {
	minetest.RegisterPacketPre(func(c *minetest.Client, cmd mt.Cmd) bool {
		var skipCmd bool

		switch cmd.(type) {
		case *mt.ToCltMedia:
			skipCmd = true

		// we dont want to send those (too many of them)
		case *mt.ToCltBlkData:
			return false
		}

		var j []byte
		var err error

		if skipCmd {
			j, err = json.Marshal(MsgPacket{
				Packet: Packet{
					Type: fmt.Sprintf("%T", cmd)[4:],
					Srv:  true,
					Clt:  c.RemoteAddr().String(),
					Name: c.Name,
				},
			})
		} else {
			j, err = json.Marshal(MsgPacket{
				Packet: Packet{
					Type: fmt.Sprintf("%T", cmd)[4:],
					Srv:  true,
					Clt:  c.RemoteAddr().String(),
					Name: c.Name,
					Cmd:  cmd,
				},
			})
		}

		if err != nil {
			fmt.Println(err)
			return false
		}

		msg := "packet " + string(j)

		broadcast(msg)

		return false
	})

	minetest.RegisterPktProcessor(func(clt *minetest.Client, pkt *mt.Pkt) {
		if _, ok := pkt.Cmd.(*mt.ToSrvGotBlks); ok {
			return
		}

		j, err := json.Marshal(MsgPacket{
			Packet: Packet{
				Type: fmt.Sprintf("%T", pkt.Cmd)[4:],
				Srv:  false,
				Clt:  clt.RemoteAddr().String(),
				Name: clt.Name,
				Cmd:  pkt.Cmd,
			},
		})

		if err != nil {
			fmt.Println(err)
			return
		}

		msg := "packet " + string(j)

		broadcast(msg)
	})
}
