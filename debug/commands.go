package debug

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/inventory"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"strconv"
	"strings"
)

func init() {
	chat.RegisterChatCmd("uuid", func(c *minetest.Client, args []string) {
		chat.SendMsgf(c, mt.RawMsg, "Your UUID is %s", c.UUID)
	})

	chat.RegisterChatCmd("fullpos", func(c *minetest.Client, args []string) {
		chat.SendMsgf(c, mt.RawMsg, "Your pos is %+v", minetest.GetFullPos(c))
	})

	chat.RegisterChatCmd("dimension", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsgf(c, mt.RawMsg, "Usage: dimension <name>")
			return
		}

		dim := minetest.GetDim(args[0])
		if dim == nil {
			chat.SendMsgf(c, mt.RawMsg, "Dimension '%s' does not exist!", args[0])
			return
		}

		chat.SendMsgf(c, mt.RawMsg, "Sending you to %s (%d)!", args[0], *dim)

		pos := minetest.GetPos(c)
		pos.Dim = dim.ID

		minetest.SetPos(c, pos)
	})

	chat.RegisterChatCmd("open_dim", func(c *minetest.Client, args []string) {
		usage := func(str string, v ...any) {
			chat.SendMsgf(c, mt.RawMsg, str+"Usage: open_dim <name> <mapgen> <mapdriver>:<file>", v...)
		}

		if len(args) != 3 {
			usage("")
			return
		}

		dimName := args[0]
		gen := args[1]

		s := strings.SplitN(args[2], ":", 3)
		if len(s) != 2 {
			usage("")
			return
		}

		driver, file := s[0], s[1]

		chat.SendMsgf(c, mt.RawMsg, "Loading new dimension %s from %s using drv %s",
			dimName, file, driver,
		)

		id, err := minetest.NewDim(dimName, gen, driver, file)
		if err != nil {
			usage("Err: %s\n", err)
			return
		}

		chat.SendMsgf(c, mt.RawMsg, "Success, got ID: %d",
			id,
		)
	})

	chat.RegisterChatCmd("load_here", func(c *minetest.Client, args []string) {
		blkpos, _ := minetest.Pos2Blkpos(minetest.GetPos(c).IntPos())

		go func() {
			ack := minetest.LoadBlk(c, blkpos)
			if ack != nil {
				<-ack
			}

			chat.SendMsgf(c, mt.RawMsg, "loadedBlk at (%d, %d, %d) %s (%d)", blkpos.Pos[0], blkpos.Pos[1], blkpos.Pos[2], blkpos.Dim, blkpos.Dim)
		}()
	})

	chat.RegisterChatCmd("kickme", func(c *minetest.Client, args []string) {
		var msg string

		if len(args) != 0 {
			msg = strings.Join(args, " ")
		} else {
			msg = "You kicked yourself!"
		}

		c.Kick(mt.Custom, msg)
	})

	chat.RegisterChatCmd("config", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsgf(c, mt.RawMsg, "Usage: config <key>")
			return
		}

		v, ok := minetest.GetConfig(args[0])
		chat.SendMsgf(c, mt.RawMsg, "value: %v, %s", v, T(ok, "set", "not set"))
	})

	chat.RegisterChatCmd("savedata", func(c *minetest.Client, args []string) {
		if len(args) < 2 {
			chat.SendMsgf(c, mt.RawMsg, "Usage: savedata <field> <data>...")
			return
		}

		field := args[0]
		data := strings.Join(args[1:], " ")

		chat.SendMsgf(c, mt.RawMsg, "Setting field '%s' to '%s'", field, data)
		c.SetData(field, &minetest.ClientDataString{String: data})

		return
	})

	chat.RegisterChatCmd("give", func(c *minetest.Client, args []string) {
		if len(args) < 2 {
			chat.SendMsgf(c, mt.RawMsg, "Usage: give <item> <count>")
			return
		}

		item := args[0]

		count, err := strconv.ParseInt(args[1], 10, 33)
		if err != nil {
			chat.SendMsgf(c, mt.RawMsg, "Error parsing argument: %s", err)
			return
		}

		if count < 0 {
			chat.SendMsgf(c, mt.RawMsg, "Not yet implemented")
			return
		}

		i, ack, err := inventory.Give(c,
			&inventory.InvLocation{
				Identifier: &inventory.InvIdentifierCurrentPlayer{},
				Name:       "main",
				Stack:      -1, // auto aquire
			},
			uint16(count), item,
		)

		if err != nil {
			chat.SendMsgf(c, mt.RawMsg, "Error: %s", err)
			return
		}

		chat.SendMsgf(c, mt.RawMsg, "Waiting for ack")
		if ack != nil {
			<-ack
		}
		chat.SendMsgf(c, mt.RawMsg, "Added %d items: %s", i, err)

		return
	})

	chat.RegisterChatCmd("getdata", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsgf(c, mt.RawMsg, "Usage: getdata <field>")
			return
		}

		field := args[0]

		data, ok := c.GetData(field)
		if !ok {
			chat.SendMsgf(c, mt.RawMsg, "Field '%s' is empty!", field)
		}

		chat.SendMsgf(c, mt.RawMsg, "Getting field '%s'", field)
		d := minetest.TryClientDataString(c, field)
		if d == nil {
			chat.SendMsgf(c, mt.RawMsg, "Field '%s' is empty! (type: %T)", field, data)

			return
		}

		chat.SendMsgf(c, mt.RawMsg, "Content: %s", d.String)

		return
	})

	chat.RegisterChatCmd("cleanmapcache", func(c *minetest.Client, args []string) {
		minetest.CleanCache()

		chat.SendMsgf(c, mt.RawMsg, "cleaning map cache done")
	})

	chat.RegisterChatCmd("nodeinfo", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage nodeinfo <[name] [param0_raw] [param1] [param2] [meta] | all>", mt.RawMsg)
			return
		}

		p, pi := minetest.Pos2Blkpos(minetest.GetPos(c).IntPos())
		blk := minetest.GetBlk(p)

		argsMap := make(map[string]struct{})

		// parse arguments:
		if args[0] == "all" {
			argsMap["name"] = struct{}{}
			argsMap["param0_raw"] = struct{}{}
			argsMap["param1"] = struct{}{}
			argsMap["param2"] = struct{}{}
			argsMap["meta"] = struct{}{}
		} else {
			for _, arg := range args {
				argsMap[arg] = struct{}{}
			}
		}

		var msg string
		sblk := blk.MapBlk

		param0 := sblk.Param0[pi]
		param1 := sblk.Param1[pi]
		param2 := sblk.Param2[pi]

		for info := range argsMap {
			switch info {
			case "name":
				msg += fmt.Sprintf("Name: %s\n", minetest.NodeIdMap[param0])
				break

			case "param0_raw":
				msg += fmt.Sprintf("Param0: %d 0x%X\n", param0, param0)
				break

			case "param1":
				msg += fmt.Sprintf("Param1: %d 0x%X\n", param1, param1)
				break

			case "param2":
				msg += fmt.Sprintf("Param2: %d 0x%X\n", param2, param2)
				break

			case "meta":
				meta, ok := sblk.NodeMetas[pi]
				if !ok {
					msg += "Meta: nil\n"
				} else {
					msg += fmt.Sprintf("Meta:\n\tFields:%s\n\t\tInv (count): %d\n",
						FormatNodeMetaField(meta.Fields),
						len(meta.Inv),
					)
				}

				break
			}
		}

		chat.SendMsg(c, msg, mt.RawMsg)
	})

	chat.RegisterChatCmd("inv", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsgf(c, mt.RawMsg, "Usage: inv <name>")
			return
		}

		inv, err := inventory.GetInv(c)
		if err != nil {
			chat.SendMsgf(c, mt.RawMsg, "Error: %s", err)
			return
		}

		inv.RLock()
		defer inv.RUnlock()
		i := inv.M[args[0]]

		chat.SendMsgf(c, mt.RawMsg, "value: %+v", i)
	})
}

func FormatNodeMetaField(nmf []mt.NodeMetaField) (str string) {
	for _, field := range nmf {
		str += fmt.Sprintf("\n\t - %s: %s", field.Name, field.Value)
		if field.Private {
			str += " (private)"
		}
	}

	return
}

func T[V any](c bool, t, f V) V {
	if c {
		return t
	} else {
		return f
	}
}
