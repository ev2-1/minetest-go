package main

import (
	"plugin"

	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go"
	"log"
)

func PluginsLoaded(pl []*plugin.Plugin) {
	log.Print("pluginsLoaded func")
}

func ProcessPkt(c *minetest.Client, pkt *mt.Pkt) {
	c.Log("Packet")
}
