package help

import (
	"github.com/ev2-1/minetest-go/chat"
	"github.com/ev2-1/minetest-go/chat/help"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"strings"
	"sync"
)

var tableOnce sync.Once
var table *Table
var execMap map[int]string
var configureMap map[int]help.Help

// Ensure InitTable
func InitTable() {
	tableOnce.Do(doInitTable)
}

func doInitTable() {
	for _, cmd := range chat.ChatCmds() {
		help := help.GetHelp(cmd)

		if help == nil {
			table = table.Append(HeadEntry{
				Fields: []string{
					cmd,
					"not provided",
					"execute",
				},
			})
		} else {
			e := HeadEntry{
				Fields: []string{
					cmd,
					help.Desc,
					"configure",
				},
			}

			for _, arg := range help.Args {
				e.Entries = append(e.Entries, Entry{
					arg.Name,
					fmt.Sprintf("(%s)", arg.Type.String()),
					"",
				})
			}

			table = table.Append(e)
		}
	}

	execMap = make(map[int]string)
	configureMap = make(map[int]help.Help)

	//do the maps
	var i int = 0
	for _, he := range *table {
		i++

		for _, f := range he.Fields {
			if f == "configure" {
				lhelp := help.GetHelp(he.Fields[0])
				if lhelp == nil {
					configureMap[i] = help.Help{Name: he.Fields[0]}
				} else {
					configureMap[i] = *lhelp
				}
			}
			if f == "execute" {
				execMap[i] = he.Fields[0]
			}
		}

		i += len(he.Entries)
	}
}

func init() {
	help.RegisterHelp("helpspec", &help.Help{
		Desc: "displays a help formspec",
		Args: help.MustParseArgs(""),
	})
	chat.RegisterChatCmd("helpspec", func(clt *minetest.Client, args []string) {
		InitTable()

		//cmds := chat.ChatCmds()
		clt.Log("spec", table.Spec())
		clt.ShowSpecf(helpspec, table.Spec())

		//		clt.ShowSpecf(helpspec, "a,0,c,d,e,1,g,h,i,1,k,l,m,g")
	})
}

type Table []HeadEntry

func (t *Table) Append(e HeadEntry) *Table {
	if t == nil {
		return &Table{e}
	}
	nt := Table(append(*t, e))

	return &nt
}

func (t *Table) Spec() (s string) {
	for _, e := range *t {
		s += e.Spec()
	}

	return
}

type HeadEntry struct {
	Fields []string

	Entries []Entry
}

func (e *HeadEntry) Spec() (s string) {
	s = "s,0,"
	for _, f := range e.Fields {
		s += fmt.Sprintf("%s,", escape(f))
	}

	for _, e := range e.Entries {
		s += e.Spec()
	}

	return
}

type Entry []string

func (e Entry) Spec() (s string) {
	s += "s,1,"

	for _, e := range e {
		s += fmt.Sprintf("%s,", escape(e))
	}

	return
}

// * `minetest.formspec_escape(string)`: returns a string
//    * escapes the characters "[", "]", "\", "," and ";", which can not be used
//      in formspecs.

func escape(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "[", "\\[")
	str = strings.ReplaceAll(str, "]", "\\]")
	str = strings.ReplaceAll(str, ",", "\\,")
	str = strings.ReplaceAll(str, ";", "\\;")

	return str
}
