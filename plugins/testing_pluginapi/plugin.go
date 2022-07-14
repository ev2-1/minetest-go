package main

import (
	"plugin"

	"log"
)

func PluginsLoaded(pl []*plugin.Plugin) {
	log.Print("pluginsLoaded func")
}
