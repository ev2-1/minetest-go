package minetest

import (
	"os"
	"strings"
)

var confMap = make(map[string]string)

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
		}
	}
}
