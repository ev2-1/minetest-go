package debug

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cltpos struct {
	Pos   *minetest.ClientPos
	Clt   *minetest.Client
	dtime time.Duration
}

func makePosCh() (func(), chan *cltpos) {
	ch := make(chan *cltpos)
	var mu sync.Mutex

	ref := minetest.RegisterPosUpdater(func(c *minetest.Client, p *minetest.ClientPos, t time.Duration) {
		mu.Lock()
		defer mu.Unlock()

		//TODO: fix
		ch <- &cltpos{p, c, t}
	})

	stop := func() {
		mu.Lock()
		defer mu.Unlock()

		ref.Stop()
		close(ch)
	}

	return stop, ch
}

func init() {
	minetest.RegisterPlaceCond(func(clt *minetest.Client, i *mt.ToSrvInteract) bool {
		cd, ok := clt.GetData("disable_place")
		if !ok {
			return true
		}

		return !cd.(bool)
	})

	minetest.RegisterDigCond(func(clt *minetest.Client, i *mt.ToSrvInteract, _ time.Duration) bool {
		cd, ok := clt.GetData("disable_dig")
		if !ok {
			return true
		}

		return !cd.(bool)
	})

	chat.RegisterChatCmd("disable_dig", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: disable_dig on|off", mt.RawMsg)
			return
		}

		c.SetData("disable_dig", args[0] == "on")
	})

	chat.RegisterChatCmd("setnode", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: setnode <name>", mt.RawMsg)
			return
		}

		def := minetest.GetNodeDef(args[0])
		if def == nil {
			chat.SendMsgf(c, mt.RawMsg, "Node %s doesn't exist", args[0])
			return
		}

		minetest.SetNode(c.GetPos().IntPos(), mt.Node{Param0: def.Thing.Param0}, nil)
	})

	chat.RegisterChatCmd("blame", func(c *minetest.Client, args []string) {
		usage := func() {
			chat.SendMsg(c, "Usage: blame node|item <name>", mt.RawMsg)
		}

		if len(args) != 2 {
			usage()
			return
		}

		switch args[0] {
		case "node":
			def := minetest.GetNodeDef(args[1])
			if def == nil {
				chat.SendMsgf(c, mt.RawMsg, "Node %s doesn't exist", args[1])
				return
			}

			chat.SendMsgf(c, mt.RawMsg, "Node '%s' can be blamed on '%s'!", args[1], def.Path())
		case "item":
			def := minetest.GetItemDef(args[1])
			if def == nil {
				chat.SendMsgf(c, mt.RawMsg, "Item %s doesn't exist", args[1])
				return
			}

			chat.SendMsgf(c, mt.RawMsg, "Item '%s' can be blamed on '%s'!", args[1], def.Path())

		default:
			usage()
		}
	})

	chat.RegisterChatCmd("disable_place", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: disable_place on|off", mt.RawMsg)
			return
		}

		c.SetData("disable_place", args[0] == "on")
	})

	chat.RegisterChatCmd("logpos", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: logpos on|off", mt.RawMsg)
			return
		}

		stopCh := make(chan struct{})

		if args[0] == "on" {
			cd, ok := c.GetData("logpos")
			if ok && cd != nil {
				chat.SendMsg(c, "Already logging!", mt.NormalMsg)
				return
			}
		} else {
			cd, ok := c.GetData("logpos")
			if !ok || cd == nil {
				chat.SendMsg(c, "not logging!", mt.NormalMsg)
				return
			}

			close(cd.(chan struct{}))
			c.SetData("logpos", nil)
			return
		}

		c.SetData("logpos", stopCh)

		stop, posch := makePosCh()

		go func() {
			for {
				select {
				case pos := <-posch:
					func() {
						pos.Pos.RLock()
						defer pos.Pos.RUnlock()

						apos := pos.Pos.CurPos.Pos
						speed := minetest.Distance(pos.Pos.CurPos.Pos.Pos, pos.Pos.OldPos.Pos.Pos) / pos.dtime.Seconds()

						chat.SendMsgf(c, mt.RawMsg, "Pos: %5.1f %5.1f %5.1f : %5s @ %.2fn/s",
							apos.Pos[0], apos.Pos[1], apos.Pos[2], apos.Dim,
							speed,
						)
					}()

				case <-stopCh:
					stop()
					return
				}
			}
		}()
	})

	chat.RegisterChatCmd("nodeidmap", func(c *minetest.Client, args []string) {
		nim, _ := minetest.NodeMaps()

		var str string

		for id, name := range nim {
			if len(str) > 1000 {
				chat.SendMsg(c, str, mt.RawMsg)
				str = ""
			}

			str += fmt.Sprintf("%5d: %s\n", id, name)
		}

		chat.SendMsg(c, str, mt.RawMsg)
	})

	chat.RegisterChatCmd("getdetached", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: getdetached [name]", mt.RawMsg)
			return
		}

		rd, err := minetest.GetDetached(args[0], c)
		if err != nil {
			c.Logger.Printf("Error: %s", err)
			return
		}

		d := rd.Thing

		ack, err := d.AddClient(c)
		if err != nil {
			c.Logger.Printf("Error: %s", err)
			return
		}

		<-ack
		c.Logger.Printf("Sent DetachedInv")

	})

	chat.RegisterChatCmd("showspec", func(c *minetest.Client, args []string) {
		c.SendCmd(&mt.ToCltShowFormspec{
			Formspec: minetest.TestSpec(),
			Formname: "lol",
		})
	})

	tp := func(c *minetest.Client, args []string) {
		switch len(args) {
		case 1:
			pos := c.GetPos()
			dim := minetest.GetDim(args[0])
			if dim == nil {
				chat.SendMsgf(c, mt.SysMsg, "Dimension '%s' does not exists!", args[0])
				return
			}

			pos.Dim = dim.ID

			minetest.SetPos(c, pos, true)

		case 3, 4:
			pos := c.GetPos()

			x, err := strconv.ParseFloat(args[0], 32)
			if err != nil {
				chat.SendMsg(c, "Your Brain | <-- [ERROR HERE] |", mt.SysMsg)
				return
			}

			pos.Pos.Pos[0] = float32(x * 10)

			x, err = strconv.ParseFloat(args[1], 32)
			if err != nil {
				chat.SendMsg(c, "Your Brain | <-- [ERROR HERE] |", mt.SysMsg)
				return
			}

			pos.Pos.Pos[1] = float32(x * 10)

			x, err = strconv.ParseFloat(args[2], 32)
			if err != nil {
				chat.SendMsg(c, "Your Brain | <-- [ERROR HERE] |", mt.SysMsg)
				return
			}

			pos.Pos.Pos[2] = float32(x * 10)

			if len(args) == 4 {
				dim := minetest.GetDim(args[3])
				if dim == nil {
					chat.SendMsgf(c, mt.SysMsg, "Dimension '%s' does not exists!", args[3])
					return
				}

				pos.Dim = dim.ID
			}

			minetest.SetPos(c, pos, true)

		default:
			chat.SendMsgf(c, mt.SysMsg, "Usage: teleport <DIM> | <x> <y> <z> [DIM]")
		}
	}

	chat.RegisterChatCmd("tp", tp)
	chat.RegisterChatCmd("teleport", tp)

	chat.RegisterChatCmd("sleep", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsgf(c, mt.RawMsg, "Usage: sleep <time>")
			return
		}

		duration, err := time.ParseDuration(args[0])
		if err != nil {
			chat.SendMsgf(c, mt.RawMsg, "Err: %s", err)
			return
		}

		time.Sleep(duration)

		chat.SendMsgf(c, mt.RawMsg, "Slept for %s", duration)
	})

	chat.RegisterChatCmd("pos", func(c *minetest.Client, _ []string) {
		pos := c.GetPos()

		chat.SendMsgf(c, mt.SysMsg, "Your position: [%#v]",
			pos,
		)
	})

	chat.RegisterChatCmd("uuid", func(c *minetest.Client, args []string) {
		chat.SendMsgf(c, mt.RawMsg, "Your UUID is %s", c.UUID)
	})

	chat.RegisterChatCmd("fullpos", func(c *minetest.Client, args []string) {
		chat.SendMsgf(c, mt.RawMsg, "Your pos is %+v", c.GetFullPos())
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

		chat.SendMsgf(c, mt.RawMsg, "Sending you to %s (%d)!", dim.Name, dim.ID)

		pos := c.GetPos()
		pos.Dim = dim.ID

		minetest.SetPos(c, pos, true)
	})

	chat.RegisterChatCmd("open_dim", func(c *minetest.Client, args []string) {
		usage := func(str string, v ...any) {
			chat.SendMsgf(c, mt.RawMsg, str+"Usage: open_dim <name> <args>@<mapgen> <mapdriver>:<file>", v...)
		}

		if len(args) != 3 {
			usage("")
			return
		}

		dimName := args[0]
		s := strings.SplitN(args[1], "@", 3)
		if len(s) != 2 {
			usage("")
			return
		}

		genargs, gen := s[0], s[1]

		s = strings.SplitN(args[2], ":", 3)
		if len(s) != 2 {
			usage("")
			return
		}

		driver, file := s[0], s[1]
		if !path.IsAbs(file) {
			file = filepath.Clean(file)
			if strings.HasPrefix(file, "..") {
				if !minetest.GetConfigV("debug-allow-abs-map-paths", false) {
					chat.SendMsgf(c, mt.RawMsg, "Loading maps from absolute paths is not allowed!")
					return
				}
			}

			file = minetest.Path("maps/" + file)
		} else {
			if !minetest.GetConfigV("debug-allow-abs-map-paths", false) {
				chat.SendMsgf(c, mt.RawMsg, "Loading maps from absolute paths is not allowed!")
				return
			}
		}

		chat.SendMsgf(c, mt.RawMsg, "Loading new dimension %s from %s using drv %s",
			dimName, file, driver,
		)

		dim, err := minetest.NewDim(dimName, gen, genargs, driver, file)
		if err != nil {
			usage("Err: %s\n", err)
			return
		}

		chat.SendMsgf(c, mt.RawMsg, "Success, got ID: %d",
			dim.ID,
		)
	})

	chat.RegisterChatCmd("load_here", func(c *minetest.Client, args []string) {
		blkpos, _ := minetest.Pos2Blkpos(c.GetPos().IntPos())

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

		v, ok := minetest.GetConfig(args[0], any(0))
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

		i, ack, err := minetest.Give(c,
			&minetest.InvLocation{
				Identifier: &minetest.InvIdentifierCurrentPlayer{},
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

		p, pi := minetest.Pos2Blkpos(c.GetPos().IntPos())
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
				def := minetest.GetNodeDefID(param0)

				msg += fmt.Sprintf("Name: %s\n", def.Thing.Name)
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

		inv, err := minetest.GetInv(c)
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
