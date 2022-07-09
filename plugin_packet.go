package minetest

import (
	"github.com/anon55555/mt"

	"reflect"
	"sync"
)

type PacketHandler struct {
	Packets map[reflect.Type]bool

	Handle func(*Client, *mt.Pkt) bool
}

var packetHandlers []*PacketHandler
var packetHandlersMu sync.RWMutex

func RegisterPacketHandler(h *PacketHandler) {
	packetHandlersMu.Lock()
	defer packetHandlersMu.Unlock()

	packetHandlers = append(packetHandlers, h)
}
