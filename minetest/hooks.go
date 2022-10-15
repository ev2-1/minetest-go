package minetest

import (
	"github.com/anon55555/mt"
	"sync"
)

var (
	leaveHooks   []func(*Leave)
	leaveHooksMu sync.RWMutex
)

func RegisterLeaveHook(h func(*Leave)) {
	leaveHooksMu.Lock()
	defer leaveHooksMu.Unlock()

	leaveHooks = append(leaveHooks, h)
}

var (
	joinHooks   []func(*Client)
	joinHooksMu sync.RWMutex
)

func RegisterJoinHook(h func(*Client)) {
	joinHooksMu.Lock()
	defer joinHooksMu.Unlock()

	joinHooks = append(joinHooks, h)
}

var (
	initHooks   []func(*Client)
	initHooksMu sync.RWMutex
)

func RegisterInitHook(h func(*Client)) {
	initHooksMu.Lock()
	defer initHooksMu.Unlock()

	initHooks = append(initHooks, h)
}

var (
	tickHooks   []func()
	tickHooksMu sync.RWMutex
)

func RegisterTickHook(h func()) {
	tickHooksMu.Lock()
	defer tickHooksMu.Unlock()

	tickHooks = append(tickHooks, h)
}

var (
	physHooksLast float32
	physHooks     []func(dtime float32)
	physHooksMu   sync.RWMutex
)

func RegisterPhysTickHook(h func(dtime float32)) {
	physHooksMu.Lock()
	defer physHooksMu.Unlock()

	physHooks = append(physHooks, h)
}

var (
	pktTickHooks   []func()
	pktTickHooksMu sync.RWMutex
)

func RegisterPktTickHook(h func()) {
	pktTickHooksMu.Lock()
	defer pktTickHooksMu.Unlock()

	pktTickHooks = append(pktTickHooks, h)
}

var (
	packetPre   []func(*Client, mt.Cmd) bool
	packetPreMu sync.RWMutex
)

func RegisterPacketPre(h func(*Client, mt.Cmd) bool) {
	packetPreMu.Lock()
	defer packetPreMu.Unlock()

	packetPre = append(packetPre, h)
}

var (
	pktProcessors   []func(*Client, *mt.Pkt)
	pktProcessorsMu sync.RWMutex
)

func RegisterPktProcessor(p func(*Client, *mt.Pkt)) {
	pktProcessorsMu.Lock()
	defer pktProcessorsMu.Unlock()

	pktProcessors = append(pktProcessors, p)
}

var (
	shutdownHooks   []func()
	shutdownHooksMu sync.RWMutex
)

func RegisterShutdownHooks(p func()) {
	shutdownHooksMu.Lock()
	defer shutdownHooksMu.Unlock()

	shutdownHooks = append(shutdownHooks, p)
}

var (
	saveFileHooks   []func()
	saveFileHooksMu sync.RWMutex
)

func RegisterSaveFileHooks(p func()) {
	saveFileHooksMu.Lock()
	defer saveFileHooksMu.Unlock()

	saveFileHooks = append(saveFileHooks, p)
}
