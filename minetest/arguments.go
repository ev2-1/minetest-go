package minetest

import (
	"log"
	"os"
	"strings"
	"time"
)

var confMap = make(map[string]string)
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
		case "v":
			verbose = v == "true"
		case "tick":
			d, err := time.ParseDuration(v)
			if err != nil {
				log.Fatal("Error parsing duration \"-tick\"", err)
			}

			tickDuration = d
		}
	}
}
