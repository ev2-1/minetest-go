package main

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"
)

func ProcessPkt(c *minetest.Client, pkt *mt.Pkt) {
	switch cmd := pkt.Cmd.(type) {
	case *mt.ToSrvPlayerPos:
		updatePos(c, &cmd.Pos)
	}
}
