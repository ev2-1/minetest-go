package minetest

import (
	"os"
	"strings"
)

var confMap = make(map[string]string)
var pluginNoLoad []string
var verbose bool = true

func parseArguments() {
	var key string
	for _, kv := range os.Args {
		if strings.HasPrefix(kv, "-") {
			key = kv[1:]
		} else {
			confMap[key] = kv
		}
	}

	// parse
	for k, v := range confMap {
		switch k {
		case "listen":
			listenAddr = v
		case "ign-plugin":
			pluginNoLoad = strings.Split(v, ",")
		case "v":
			verbose = v == "true"
		}
	}
}

func pluginNotLoad(file string) bool {
	for _, p := range pluginNoLoad {
		if p == file || p+".so" == file {
			return true
		}
	}

	return false
}
