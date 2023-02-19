package minetest

import (
	"github.com/anon55555/mt"

	"fmt"
	"runtime"
	"sync"
)

type Registerd[K any] struct {
	Path  string
	Thing K
}

// Returns file:line of caller at i
// with 0 identifying the caller of Path
func Caller(i int) string {
	_, file, line, _ := runtime.Caller(i + 1)
	return fmt.Sprintf("%s:%d", file, line)
}

type LeaveHook func(*Leave)

var (
	leaveHooks   []Registerd[LeaveHook]
	leaveHooksMu sync.RWMutex
)

// Gets called after client has left (*Client struct still exists)
func RegisterLeaveHook(h LeaveHook) {
	leaveHooksMu.Lock()
	defer leaveHooksMu.Unlock()

	leaveHooks = append(leaveHooks, Registerd[LeaveHook]{Caller(1), h})
}

type JoinHook func(*Client)

var (
	joinHooks   []Registerd[JoinHook]
	joinHooksMu sync.RWMutex
)

// Gets called after client is initialized
// After player is given controll
func RegisterJoinHook(h JoinHook) {
	joinHooksMu.Lock()
	defer joinHooksMu.Unlock()

	joinHooks = append(joinHooks, Registerd[JoinHook]{Caller(1), h})
}

type RegisterHook func(*Client)

var (
	registerHooks   []Registerd[RegisterHook]
	registerHooksMu sync.RWMutex
)

// Gets called at registrationtime
// intended to initialize client data
func RegisterRegisterHook(h RegisterHook) {
	registerHooksMu.Lock()
	defer registerHooksMu.Unlock()

	registerHooks = append(registerHooks, Registerd[RegisterHook]{Caller(1), h})
}

type InitHook func(*Client)

var (
	initHooks   []Registerd[InitHook]
	initHooksMu sync.RWMutex
)

// Gets called as soon as client authentication is successfull
func RegisterInitHook(h InitHook) {
	initHooksMu.Lock()
	defer initHooksMu.Unlock()

	initHooks = append(initHooks, Registerd[InitHook]{Caller(1), h})
}

type TickHook func()

var (
	tickHooks   []Registerd[TickHook]
	tickHooksMu sync.RWMutex
)

// Gets called each tick
func RegisterTickHook(h func()) {
	tickHooksMu.Lock()
	defer tickHooksMu.Unlock()

	tickHooks = append(tickHooks, Registerd[TickHook]{Caller(1), h})
}

type PhysHook func(dtime float32)

var (
	physHooksLast float32
	physHooks     []Registerd[PhysHook]
	physHooksMu   sync.RWMutex
)

// Gets called each tick
// dtime is time since last tick
func RegisterPhysTickHook(h PhysHook) {
	physHooksMu.Lock()
	defer physHooksMu.Unlock()

	physHooks = append(physHooks, Registerd[PhysHook]{Caller(1), h})
}

type PktTickHook func()

var (
	pktTickHooks   []Registerd[PktTickHook]
	pktTickHooksMu sync.RWMutex
)

// Gets called at end of each tick
// If you can, send packets in here
func RegisterPktTickHook(h PktTickHook) {
	pktTickHooksMu.Lock()
	defer pktTickHooksMu.Unlock()

	pktTickHooks = append(pktTickHooks, Registerd[PktTickHook]{Caller(1), h})
}

type PacketPre func(*Client, mt.Cmd) bool

var (
	packetPre   []Registerd[PacketPre]
	packetPreMu sync.RWMutex
)

// Gets called before packet reaches Processors
// If (one) func returns false packet is dropped
func RegisterPacketPre(h PacketPre) {
	packetPreMu.Lock()
	defer packetPreMu.Unlock()

	packetPre = append(packetPre, Registerd[PacketPre]{Caller(1), h})
}

type PktProcessor func(*Client, *mt.Pkt)

var (
	pktProcessors   []Registerd[PktProcessor]
	pktProcessorsMu sync.RWMutex
)

// Gets called for each packet received
func RegisterPktProcessor(h PktProcessor) {
	pktProcessorsMu.Lock()
	defer pktProcessorsMu.Unlock()

	pktProcessors = append(pktProcessors, Registerd[PktProcessor]{Caller(1), h})
}

type ShutdownHook func()

var (
	shutdownHooks   []Registerd[ShutdownHook]
	shutdownHooksMu sync.RWMutex
)

// Gets called when server shuts down
// NOTE: (Leave hooks also get called)
func RegisterShutdownHook(h ShutdownHook) {
	shutdownHooksMu.Lock()
	defer shutdownHooksMu.Unlock()

	shutdownHooks = append(shutdownHooks, Registerd[ShutdownHook]{Caller(1), h})
}

type SaveFileHook func()

var (
	saveFileHooks   []Registerd[SaveFileHook]
	saveFileHooksMu sync.RWMutex
)

func RegisterSaveFileHook(h SaveFileHook) {
	saveFileHooksMu.Lock()
	defer saveFileHooksMu.Unlock()

	saveFileHooks = append(saveFileHooks, Registerd[SaveFileHook]{Caller(1), h})
}
