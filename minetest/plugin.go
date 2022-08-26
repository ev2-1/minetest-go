package minetest

import (
	"log"
	"os"
	"plugin"
	"sync"
)

var pluginsOnce sync.Once
var plugins = make(map[string]*plugin.Plugin)

var loadingPlugin string

func loadPlugins() {
	pluginsOnce.Do(func() {
		path := Path("plugins")
		os.Mkdir(path, 0777)

		files, err := os.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}

		//var plugins []*plugin.Plugin
		var loader = make(map[string]func(map[string]*plugin.Plugin))

		for _, file := range files {
			loadingPlugin = file.Name()
			if loadingPlugin[0] == "."[0] {
				continue
			}

			if pluginNotLoad(file.Name()) {
				log.Print("[plugins] skipping ", file.Name())
				continue
			}
			log.Print("[plugins] loading ", file.Name())

			p, err := plugin.Open(path + "/" + file.Name())
			if err != nil {
				log.Print(err)
				continue
			}

			pname := file.Name()

			n, err := p.Lookup("Name")
			if err == nil {
				name, ok := n.(*string)
				if ok {
					pname = *name
				}
			}

			plugins[pname] = p

			l, err := p.Lookup("PluginsLoaded")
			if err == nil {
				switch lo := l.(type) {
				case func(map[string]*plugin.Plugin):
					loader[pname] = lo
				}
			}
		}

		loadingPlugin = ""
		log.Print("[plugins] loading done")

		log.Print("[media] filling NameID map")
		FillNameIdMap()
		log.Print("[media] filling NameID map done")

		log.Print("[plugins] PluginsLoaded hooks")
		for name, l := range loader {
			log.Print("[plugins] PluginsLoaded - ", name)
			l(plugins)
		}
		pluginHook(plugins)

		log.Print("[plugins] PluginsLoaded hooks done")

	})
}
