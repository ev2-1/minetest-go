package help

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"strconv"
	"strings"
	"time"
)

var helpspec = &minetest.Formspec{
	Name: "helpspec",
	Spec: `formspec_version[6]
size[13,6.5]
label[0,-0.1;` + escape("Help menu:") + `]
tablecolumns[color;tree;text;text;text]
table[0,0.5;12.8,5.5;list;%s;0]
button_exit[5,6;3,1;quit;Quit]`,
	Submit: helpSpecSubmit,
}

func helpSpecSubmit(c *minetest.Client, values map[string]string, time time.Duration, closed bool) {
	InitTable()
	chat.SendMsgf(c, mt.NormalMsg, "values submitted: %+#v", values)

	if closed || values["quit"] != "" || values["list"] == "" {
		return
	}

	// values["list"] e.g. CHG:11:4
	split := strings.Split(values["list"], ":")
	if len(split) != 3 {
		return
	}

	act, row, col := split[0], split[1], split[2]
	if act != "DCL" || col != "5" {
		return
	}

	i, err := strconv.Atoi(row)
	if err != nil {
		return
	}

	cmd, ok := execMap[i]
	if ok {
		c.ShowSpecf(draftspec, cmd)
		//TODO: execute

		chat.SendMsgf(c, mt.NormalMsg, "Executing /%s", cmd)
		return
	}

	cmdhelp, ok := configureMap[i]
	if ok {
		cmd = cmdhelp.Name
		for _, arg := range cmdhelp.Args {
			cmd += fmt.Sprintf(" <%s (%s)>", arg.Name, arg.Type)
		}

		c.ShowSpecf(draftspec, escape(cmd))

		chat.SendMsgf(c, mt.NormalMsg, "configuring /%s", cmdhelp.Name)
		return
	}
}

var draftspec = &minetest.Formspec{
	Name: "draftspec",
	Spec: `formspec_version[6]
size[10.5,4]
label[0.3,0.5;Draft Execute]
field[0.6,1;9.2,0.8;command;;%s]
button_exit[6.8,2.4;3,0.8;exit;Exit]
button_exit[0.6,2.4;3,0.8;execute;Execute]`,
	Submit: draftSpecSubmit,
}

func draftSpecSubmit(c *minetest.Client, values map[string]string, time time.Duration, closed bool) {
	//Submit empty spec
	if values["quit"] != "true" || values["command"] == "" {
		return
	}

	//execute command:
	exec := values["command"]
	chat.HandleCmd(c, exec)
}
