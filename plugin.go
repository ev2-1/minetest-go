package minetest

import (
	"log"
	"os"
	"plugin"
	"reflect"
	"sync"
)

var pluginsOnce sync.Once
var plugins []*plugin.Plugin

var loadingPlugin string

var (
	pluginLoadFuncType = reflect.TypeOf(func([]*plugin.Plugin) {})
)

func loadPlugins() {
	pluginsOnce.Do(func() {
		path := Path("plugins")
		os.Mkdir(path, 0777)

		files, err := os.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}

		//var plugins []*plugin.Plugin
		var loader []func([]*plugin.Plugin)

		for _, file := range files {
			loadingPlugin = file.Name()
			if loadingPlugin[0] == "."[0] {
				continue
			}

			log.Print("loading plugin ", file.Name())

			p, err := plugin.Open(path + "/" + file.Name())
			if err != nil {
				log.Print(err)
				continue
			}

			plugins = append(plugins, p)

			l, err := p.Lookup("PluginsLoaded")
			if err == nil {
				switch lo := l.(type) {
				case func([]*plugin.Plugin):
					loader = append(loader, lo)
				}
			}

		}

		loadingPlugin = ""
		log.Print("loaded plugins")

		log.Print("after loaded plugins")
		for _, l := range loader {
			l(plugins)
		}

		log.Print("loaded plugins complete")

	})
}
