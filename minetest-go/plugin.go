package minetest

import (
	"log"
	"os"
	"plugin"
	"sync"
)

var pluginsOnce sync.Once

var loadingPlugin string

func loadPlugins() {
	pluginsOnce.Do(func() {
		path := Path("plugins")
		os.Mkdir(path, 0777)

		files, err := os.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			loadingPlugin = file.Name()

			log.Print("loading plugin ", file.Name())

			_, err := plugin.Open(path + "/" + file.Name())
			if err != nil {
				log.Print(err)
				continue
			}
		}

		loadingPlugin = ""

		log.Print("load plugins")
	})
}
