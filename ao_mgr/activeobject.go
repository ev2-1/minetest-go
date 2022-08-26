package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"sync"
)

const (
	GlobalAOIDmax = mt.AOID(2 ^ 32)
	GlobalAOIDmin = mt.AOID(1)
)

var aosMu sync.RWMutex
var aos = make(map[mt.AOID]aoidType)
var aosClt = make(map[mt.AOID]map[*minetest.Client]struct{})

type aoidType bool

const (
	Client aoidType = true
	Global          = false
)

// TODO clear all client aoids
