package debug

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/ao_mgr"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/chat/help"
	"github.com/ev2-1/minetest-go/minetest"

	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func Encode(v any) string {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("| ", "  ")

	err := enc.Encode(v)
	if err != nil {
		return "Error encoding: " + err.Error()
	}

	return "| " + buf.String() + "^^^"
}

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

	help.RegisterHelp("info", &help.Help{
		Desc: "prints information about node- and itemdefs",
		Args: help.MustParseArgs("string:mode string:name"),
	})
	chat.RegisterChatCmd("info", func(c *minetest.Client, args []string) {
		usage := func() {
			chat.SendMsg(c, "Usage: info node|item|mt_item <name>", mt.RawMsg)
		}

		if len(args) != 2 {
			usage()
			return
		}

		switch args[0] {
		case "node":
			def := minetest.GetNodeDef(args[1])
			if def == nil {
				chat.SendMsgf(c, mt.RawMsg, "Node '%s' doesn't exist", args[1])
				return
			}

			chat.SendMsgf(c, mt.RawMsg, "Registerd By %s\nDef:\n%s", def.Path(), Encode(def.Thing))

		case "item":
			def := minetest.GetItemDef(args[1])
			if def == nil {
				chat.SendMsgf(c, mt.RawMsg, "Item %s doesn't exist", args[1])
				return
			}

			chat.SendMsgf(c, mt.RawMsg, "Registerd By %s\nDef:\n%s", def.Path(), Encode(def.Thing))

		case "mt_item":
			def := minetest.GetItemDef(args[1])
			if def == nil {
				chat.SendMsgf(c, mt.RawMsg, "Item %s doesn't exist", args[1])
				return
			}

			chat.SendMsgf(c, mt.RawMsg, "Registerd By %s\nDef:\n%s", def.Path(), Encode(def.Thing.ItemDef()))

		default:
			usage()
		}
	})

	help.RegisterHelp("disable_dig", &help.Help{
		Desc: "disables digging for self though a DigCondition",
		Args: help.MustParseArgs("bool:toggle"),
	})
	chat.RegisterChatCmd("disable_dig", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: disable_dig on|off", mt.RawMsg)
			return
		}

		c.SetData("disable_dig", args[0] == "on")
	})

	help.RegisterHelp("setnode", &help.Help{
		Desc: "sets node in which player stands to given name",
		Args: help.MustParseArgs("string:name"),
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

	help.RegisterHelp("blame", &help.Help{
		Desc: "prints the line of code a given node or item was registered",
		Args: help.MustParseArgs("string:type string:name"),
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

	help.RegisterHelp("disable_place", &help.Help{
		Desc: "disables the placing of nodes for player",
		Args: help.MustParseArgs("string:toggle"),
	})
	chat.RegisterChatCmd("disable_place", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: disable_place on|off", mt.RawMsg)
			return
		}

		c.SetData("disable_place", args[0] == "on")
	})

	help.RegisterHelp("logpos", &help.Help{
		Desc: "toggles logging of client pos to chat",
		Args: help.MustParseArgs("string:toggle"),
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
						speed := minetest.Distance(pos.Pos.CurPos.Pos.Pos, pos.Pos.OldPos.Pos.Pos) / float32(pos.dtime.Seconds())

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

	help.RegisterHelp("nodeidmap", &help.Help{
		Desc: "prints nodeidmap",
		Args: help.MustParseArgs(""),
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

	help.RegisterHelp("getdetached", &help.Help{
		Desc: "sends client detached inventory",
		Args: help.MustParseArgs("string:name"),
	})
	chat.RegisterChatCmd("getdetached", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsg(c, "Usage: getdetached [name]", mt.RawMsg)
			return
		}

		rd, err := minetest.GetDetached(args[0], c)
		if err != nil {
			c.Logf("Error: %s", err)
			return
		}

		d := rd.Thing

		ack, err := d.AddClient(c)
		if err != nil {
			c.Logf("Error: %s", err)
			return
		}

		<-ack
		c.Log("Sent DetachedInv")
	})

	help.RegisterHelp("showspec", &help.Help{
		Desc: "shows client a test formspec",
		Args: help.MustParseArgs(""),
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

	help.RegisterHelp("tp", &help.Help{
		Desc: "short command for teleport",
		Args: help.MustParseArgs("pos:destination string:dimension"),
	})
	chat.RegisterChatCmd("tp", tp)
	help.RegisterHelp("teleport", &help.Help{
		Desc: "teleports player",
		Args: help.MustParseArgs("pos:destination string:dimension"),
	})
	chat.RegisterChatCmd("teleport", tp)

	help.RegisterHelp("sleep", &help.Help{
		Desc: "sleeps before returning; time is parsed using time.ParseDuration",
		Args: help.MustParseArgs("string:duration"),
	})
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

	help.RegisterHelp("pos", &help.Help{
		Desc: "pos prints current PPos",
		Args: help.MustParseArgs(""),
	})
	chat.RegisterChatCmd("pos", func(c *minetest.Client, _ []string) {
		pos := c.GetPos()

		chat.SendMsgf(c, mt.SysMsg, "Your position: [%#v]",
			pos,
		)
	})

	help.RegisterHelp("uuid", &help.Help{
		Desc: "prints your uuid",
		Args: help.MustParseArgs(""),
	})
	chat.RegisterChatCmd("uuid", func(c *minetest.Client, args []string) {
		chat.SendMsgf(c, mt.RawMsg, "Your UUID is %s", c.UUID)
	})

	help.RegisterHelp("fullpos", &help.Help{
		Desc: "prints full client position",
		Args: help.MustParseArgs(""),
	})
	chat.RegisterChatCmd("fullpos", func(c *minetest.Client, args []string) {
		chat.SendMsgf(c, mt.RawMsg, "Your pos is %+v", c.GetFullPos())
	})

	help.RegisterHelp("fix_aos", &help.Help{
		Desc: "resends all ActiveObjects",
		Args: help.MustParseArgs(""),
	})
	chat.RegisterChatCmd("fix_aos", func(c *minetest.Client, args []string) {
		cd := ao.GetClientData(c)
		if cd == nil {
			chat.SendMsgf(c, mt.RawMsg, "you don't have ao.ClientData!")
			return
		}

		cd.Lock()
		defer cd.Unlock()

		// reset map
		cd.AOs = make(map[mt.AOID]struct{})
		chat.SendMsgf(c, mt.RawMsg, "reset cd.AOs!")
	})

	help.RegisterHelp("list_aos", &help.Help{
		Desc: "lists all ActiveObjects in space",
		Args: help.MustParseArgs("string:space"),
	})
	chat.RegisterChatCmd("list_aos", func(c *minetest.Client, args []string) {
		usage := func() { chat.SendMsg(c, "usage: list_aos local|global", mt.RawMsg) }

		if len(args) != 1 {
			usage()
			return
		}

		switch args[0] {
		case "local":
			cd := ao.GetClientData(c)
			cd.RLock()
			aos := make([]mt.AOID, 0, len(cd.AOs))
			for k := range cd.AOs {
				aos = append(aos, k)
			}

			cd.RUnlock()

			chat.SendMsgf(c, mt.RawMsg, "you locally have the following ids: %v", aos)

		case "global":
			aos := make([]mt.AOID, 0)

			for id := range ao.ListAOs() {
				aos = append(aos, id)
			}

			chat.SendMsgf(c, mt.RawMsg, "there are these ids globally: %v", aos)

		default:
			usage()
		}
	})

	help.RegisterHelp("dimension", &help.Help{
		Desc: "sends client to specified dimension",
		Args: help.MustParseArgs("string:dimension"),
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

	help.RegisterHelp("open_dim", &help.Help{
		Desc: "opens dimension with parameters",
		Args: help.MustParseArgs("string:name string:mapgen_arguments string:mapdriver_arguments"),
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

	help.RegisterHelp("load_here", &help.Help{
		Desc: "loads mapblock corosponding to clients position",
		Args: help.MustParseArgs(""),
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

	help.RegisterHelp("kickme", &help.Help{
		Desc: "kicks yourself",
		Args: help.MustParseArgs("string:reason"),
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

	help.RegisterHelp("config", &help.Help{
		Desc: "reads out config value",
		Args: help.MustParseArgs("string:key"),
	})
	chat.RegisterChatCmd("config", func(c *minetest.Client, args []string) {
		if len(args) != 1 {
			chat.SendMsgf(c, mt.RawMsg, "Usage: config <key>")
			return
		}

		v, ok := minetest.GetConfig(args[0], any(0))
		chat.SendMsgf(c, mt.RawMsg, "value: %v, %s", v, T(ok, "set", "not set"))
	})

	help.RegisterHelp("aoid", &help.Help{
		Desc: "prints aoid nelonging to playerAO",
		Args: help.MustParseArgs(""),
	})
	chat.RegisterChatCmd("aoid", func(c *minetest.Client, args []string) {
		chat.SendMsgf(c, mt.RawMsg, "Your AOID is %d", ao.GetPAOID(c))
	})

	help.RegisterHelp("savedata", &help.Help{
		Desc: "modiefies clientdata",
		Args: help.MustParseArgs("string:key string:value"),
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

	help.RegisterHelp("give", &help.Help{
		Desc: "adds items to inventory",
		Args: help.MustParseArgs("string:item int:count"),
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

	help.RegisterHelp("getdata", &help.Help{
		Desc: "prints clientdata",
		Args: help.MustParseArgs("string:key"),
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

	help.RegisterHelp("cleanmapcache", &help.Help{
		Desc: "triggers map cache clean",
		Args: help.MustParseArgs(""),
	})
	chat.RegisterChatCmd("cleanmapcache", func(c *minetest.Client, args []string) {
		minetest.CleanCache()

		chat.SendMsgf(c, mt.RawMsg, "cleaning map cache done")
	})

	help.RegisterHelp("nodeinfo", &help.Help{
		Desc: "prints node information client is standing in",
		Args: help.MustParseArgs("any:arguments"),
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

	help.RegisterHelp("inv", &help.Help{
		Desc: "prints clients inv with specified name",
		Args: help.MustParseArgs("string:name"),
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

	help.RegisterHelp("help", &help.Help{
		Desc: "prints basic help menu",
		Args: help.MustParseArgs(""),
	})
	chat.RegisterChatCmd("help", func(c *minetest.Client, args []string) {
		if len(args) == 0 {
			texts := HelpTexts()
			s := make([]string, len(texts))
			var i int
			for name, desc := range texts {
				s[i] = fmt.Sprintf("%s: \n %s", name, strings.ReplaceAll(desc, "\n", "\n "))

				i++
			}

			chat.SendMsg(c, strings.Join(s, "\n"), mt.NormalMsg)
		} else {
			h := help.GetHelp(args[0])
			if h == nil {
				chat.SendMsgf(c, mt.NormalMsg, "%s has no help")
			} else {
				chat.SendMsgf(c, mt.NormalMsg, "%s:\n %s arguments:\n  %s",
					h.Name,
					h.Desc,
					strings.ReplaceAll(Args(h.Args), "\n", "\n  "),
				)
			}
		}
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

func HelpTexts() map[string]string {
	cmds := chat.ChatCmds()
	m := make(map[string]string, len(cmds))

	for _, cmd := range cmds {
		if h := help.GetHelp(cmd); h == nil {
			m[cmd] = "no help provided"
		} else {
			m[cmd] = fmt.Sprintf("%s\nusage:\n %s", h.Desc,
				strings.ReplaceAll(Args(h.Args), "\n", "\n  "))
		}
	}

	return m
}

func Args(args []help.Argument) (s string) {
	if args == nil || len(args) == 0 {
		return "none provided"
	}

	for _, arg := range args {
		s += fmt.Sprintf("%s - %s\n", arg.Name, arg.Type)
	}

	return s[:len(s)-1]
}
