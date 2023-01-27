package minetest

import (
	"github.com/anon55555/mt"
	"sync"
)

var (
	leaveHooks   []func(*Leave)
	leaveHooksMu sync.RWMutex
)

// Gets called after client has left (*Client struct still exists)
func RegisterLeaveHook(h func(*Leave)) {
	leaveHooksMu.Lock()
	defer leaveHooksMu.Unlock()

	leaveHooks = append(leaveHooks, h)
}

var (
	joinHooks   []func(*Client)
	joinHooksMu sync.RWMutex
)

// Gets called after client is initialized
// After player is given controll
func RegisterJoinHook(h func(*Client)) {
	joinHooksMu.Lock()
	defer joinHooksMu.Unlock()

	joinHooks = append(joinHooks, h)
}

var (
	registerHooks   []func(*Client)
	registerHooksMu sync.RWMutex
)

// Gets called at registrationtime
// intended to initialize client data
func RegisterRegisterHook(h func(*Client)) {
	registerHooksMu.Lock()
	defer registerHooksMu.Unlock()

	registerHooks = append(registerHooks, h)
}

var (
	initHooks   []func(*Client)
	initHooksMu sync.RWMutex
)

// Gets called as soon as client authentication is successfull
func RegisterInitHook(h func(*Client)) {
	initHooksMu.Lock()
	defer initHooksMu.Unlock()

	initHooks = append(initHooks, h)
}

var (
	tickHooks   []func()
	tickHooksMu sync.RWMutex
)

// Gets called each tick
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

// Gets called each tick
// dtime is time since last tick
func RegisterPhysTickHook(h func(dtime float32)) {
	physHooksMu.Lock()
	defer physHooksMu.Unlock()

	physHooks = append(physHooks, h)
}

var (
	pktTickHooks   []func()
	pktTickHooksMu sync.RWMutex
)

// Gets called at end of each tick
// If you can, send packets in here
func RegisterPktTickHook(h func()) {
	pktTickHooksMu.Lock()
	defer pktTickHooksMu.Unlock()

	pktTickHooks = append(pktTickHooks, h)
}

var (
	packetPre   []func(*Client, mt.Cmd) bool
	packetPreMu sync.RWMutex
)

// Gets called before packet reaches Processors
// If (one) func returns false packet is dropped
func RegisterPacketPre(h func(*Client, mt.Cmd) bool) {
	packetPreMu.Lock()
	defer packetPreMu.Unlock()

	packetPre = append(packetPre, h)
}

var (
	pktProcessors   []func(*Client, *mt.Pkt)
	pktProcessorsMu sync.RWMutex
)

// Gets called for each packet received
func RegisterPktProcessor(p func(*Client, *mt.Pkt)) {
	pktProcessorsMu.Lock()
	defer pktProcessorsMu.Unlock()

	pktProcessors = append(pktProcessors, p)
}

var (
	shutdownHooks   []func()
	shutdownHooksMu sync.RWMutex
)

// Gets called when server shuts down
// NOTE: (Leave hooks also get called)
func RegisterShutdownHook(p func()) {
	shutdownHooksMu.Lock()
	defer shutdownHooksMu.Unlock()

	shutdownHooks = append(shutdownHooks, p)
}

var (
	saveFileHooks   []func()
	saveFileHooksMu sync.RWMutex
)

func RegisterSaveFileHook(p func()) {
	saveFileHooksMu.Lock()
	defer saveFileHooksMu.Unlock()

	saveFileHooks = append(saveFileHooks, p)
}
