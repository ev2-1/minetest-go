package debug

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	mmap "github.com/ev2-1/minetest-go/map"
	"github.com/ev2-1/minetest-go/minetest"
	"github.com/ev2-1/minetest-go/tools/pos"

	"fmt"
)

func init() {
	chat.RegisterChatCmd("load_here", func(c *minetest.Client, args []string) {
		blkpos, _ := mt.Pos2Blkpos(pos.GetPos(c).Pos().Int())

		<-mmap.LoadBlk(c, blkpos)

		chat.SendMsgf(c, mt.RawMsg, "loadedBlk at (%d, %d, %d)", blkpos[0], blkpos[1], blkpos[2])
	})

	chat.RegisterChatCmd("cleanmapcache", func(c *minetest.Client, args []string) {
		mmap.CleanCache()

		chat.SendMsgf(c, mt.RawMsg, "cleaning map cache done")
	})

	chat.RegisterChatCmd("nodeinfo", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage nodeinfo <[name] [param0_raw] [param1] [param2] [meta] | all>", mt.RawMsg)
			return
		}

		p, pi := mt.Pos2Blkpos(pos.GetPos(c).Pos().Int())
		blk := mmap.GetBlk(p)

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
		param0 := blk.Param0[pi]
		param1 := blk.Param1[pi]
		param2 := blk.Param2[pi]

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
				meta, ok := blk.NodeMetas[pi]
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
