package minetest

import (
	"github.com/anon55555/mt"

	"fmt"
	"runtime"
	"sync"
)

type HookRef[H any] struct {
	mapMu  *sync.RWMutex
	mapRef map[*H]struct{}
	ref    *H
}

// i: index; l: length of removal
func splice[K any](s []K, i, l int) []K {
	return append(s[:i], s[i+l:]...)
}

// Remove HookRef from Hooks
func (hr *HookRef[H]) Stop() {
	hr.mapMu.Lock()
	defer hr.mapMu.Unlock()

	delete(hr.mapRef, hr.ref)
}

type Registerd[T any] struct {
	Path  string
	Thing T
}

// Returns file:line of caller at i
// with 0 identifying the caller of Path
func Caller(i int) string {
	_, file, line, _ := runtime.Caller(i + 1)
	return fmt.Sprintf("%s:%d", file, line)
}

type LeaveHook func(*Leave)

var (
	leaveHooks   = make(map[*Registerd[LeaveHook]]struct{})
	leaveHooksMu sync.RWMutex
)

// Gets called after client has left (*Client struct still exists)
func RegisterLeaveHook(h LeaveHook) HookRef[Registerd[LeaveHook]] {
	leaveHooksMu.Lock()
	defer leaveHooksMu.Unlock()

	r := &Registerd[LeaveHook]{Caller(1), h}
	ref := HookRef[Registerd[LeaveHook]]{&leaveHooksMu, leaveHooks, r}

	leaveHooks[r] = struct{}{}

	return ref
}

type JoinHook func(*Client)

var (
	joinHooks   = make(map[*Registerd[JoinHook]]struct{})
	joinHooksMu sync.RWMutex
)

// Gets called after client is initialized
// After player is given controll
func RegisterJoinHook(h JoinHook) HookRef[Registerd[JoinHook]] {
	joinHooksMu.Lock()
	defer joinHooksMu.Unlock()

	r := &Registerd[JoinHook]{Caller(1), h}
	ref := HookRef[Registerd[JoinHook]]{&joinHooksMu, joinHooks, r}

	joinHooks[r] = struct{}{}

	return ref
}

type RegisterHook func(*Client)

var (
	registerHooks   = make(map[*Registerd[RegisterHook]]struct{})
	registerHooksMu sync.RWMutex
)

// Gets called at registrationtime
// intended to initialize client data
func RegisterRegisterHook(h RegisterHook) HookRef[Registerd[RegisterHook]] {
	registerHooksMu.Lock()
	defer registerHooksMu.Unlock()

	r := &Registerd[RegisterHook]{Caller(1), h}
	ref := HookRef[Registerd[RegisterHook]]{&registerHooksMu, registerHooks, r}

	registerHooks[r] = struct{}{}

	return ref
}

type InitHook func(*Client)

var (
	initHooks   = make(map[*Registerd[InitHook]]struct{})
	initHooksMu sync.RWMutex
)

// Gets called as soon as client authentication is successfull
func RegisterInitHook(h InitHook) HookRef[Registerd[InitHook]] {
	initHooksMu.Lock()
	defer initHooksMu.Unlock()

	r := &Registerd[InitHook]{Caller(1), h}
	ref := HookRef[Registerd[InitHook]]{&initHooksMu, initHooks, r}

	initHooks[r] = struct{}{}

	return ref
}

type TickHook func()

var (
	tickHooks   = make(map[*Registerd[TickHook]]struct{})
	tickHooksMu sync.RWMutex
)

// Gets called each tick
func RegisterTickHook(h func()) HookRef[Registerd[TickHook]] {
	tickHooksMu.Lock()
	defer tickHooksMu.Unlock()

	r := &Registerd[TickHook]{Caller(1), h}
	ref := HookRef[Registerd[TickHook]]{&tickHooksMu, tickHooks, r}

	tickHooks[r] = struct{}{}

	return ref
}

type PhysHook func(dtime float32)

var (
	physHooksLast float32
	physHooks     = make(map[*Registerd[PhysHook]]struct{})
	physHooksMu   sync.RWMutex
)

// Gets called each tick
// dtime is time since last tick
func RegisterPhysTickHook(h PhysHook) HookRef[Registerd[PhysHook]] {
	physHooksMu.Lock()
	defer physHooksMu.Unlock()

	r := &Registerd[PhysHook]{Caller(1), h}
	ref := HookRef[Registerd[PhysHook]]{&physHooksMu, physHooks, r}

	physHooks[r] = struct{}{}

	return ref
}

type PktTickHook func()

var (
	pktTickHooks   = make(map[*Registerd[PktTickHook]]struct{})
	pktTickHooksMu sync.RWMutex
)

// Gets called at end of each tick
// If you can, send packets in here
func RegisterPktTickHook(h PktTickHook) HookRef[Registerd[PktTickHook]] {
	pktTickHooksMu.Lock()
	defer pktTickHooksMu.Unlock()

	r := &Registerd[PktTickHook]{Caller(1), h}
	ref := HookRef[Registerd[PktTickHook]]{&pktTickHooksMu, pktTickHooks, r}

	pktTickHooks[r] = struct{}{}

	return ref
}

type PacketPre func(*Client, mt.Cmd) bool

var (
	packetPre   = make(map[*Registerd[PacketPre]]struct{})
	packetPreMu sync.RWMutex
)

// Gets called before packet reaches Processors
// If (one) func returns false packet is dropped
func RegisterPacketPre(h PacketPre) HookRef[Registerd[PacketPre]] {
	packetPreMu.Lock()
	defer packetPreMu.Unlock()

	r := &Registerd[PacketPre]{Caller(1), h}
	ref := HookRef[Registerd[PacketPre]]{&packetPreMu, packetPre, r}

	packetPre[r] = struct{}{}

	return ref
}

type PktProcessor func(*Client, *mt.Pkt)

var (
	pktProcessors   = make(map[*Registerd[PktProcessor]]struct{})
	pktProcessorsMu sync.RWMutex
)

// Gets called for each packet received
func RegisterPktProcessor(h PktProcessor) HookRef[Registerd[PktProcessor]] {
	pktProcessorsMu.Lock()
	defer pktProcessorsMu.Unlock()

	r := &Registerd[PktProcessor]{Caller(1), h}
	ref := HookRef[Registerd[PktProcessor]]{&pktProcessorsMu, pktProcessors, r}

	pktProcessors[r] = struct{}{}

	return ref
}

type ShutdownHook func()

var (
	shutdownHooks   = make(map[*Registerd[ShutdownHook]]struct{})
	shutdownHooksMu sync.RWMutex
)

// Gets called when server shuts down
// NOTE: (Leave hooks also get called)
func RegisterShutdownHook(h ShutdownHook) HookRef[Registerd[ShutdownHook]] {
	shutdownHooksMu.Lock()
	defer shutdownHooksMu.Unlock()

	r := &Registerd[ShutdownHook]{Caller(1), h}
	ref := HookRef[Registerd[ShutdownHook]]{&shutdownHooksMu, shutdownHooks, r}

	shutdownHooks[r] = struct{}{}

	return ref
}

type SaveFileHook func()

var (
	saveFileHooks   = make(map[*Registerd[SaveFileHook]]struct{})
	saveFileHooksMu sync.RWMutex
)

func RegisterSaveFileHook(h SaveFileHook) HookRef[Registerd[SaveFileHook]] {
	saveFileHooksMu.Lock()
	defer saveFileHooksMu.Unlock()

	r := &Registerd[SaveFileHook]{Caller(1), h}
	ref := HookRef[Registerd[SaveFileHook]]{&saveFileHooksMu, saveFileHooks, r}

	saveFileHooks[r] = struct{}{}

	return ref
}
