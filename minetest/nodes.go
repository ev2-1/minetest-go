package minetest

import (
	"github.com/anon55555/mt"

	"sync"
	"time"
)

func doDigConds(clt *Client, i *mt.ToSrvInteract, dtime time.Duration) bool {
	digCondsMu.RLock()
	defer digCondsMu.RUnlock()

	for c := range digConds {
		if !c.Thing(clt, i, dtime) {
			return false
		}
	}

	return true
}

type DigCond func(*Client, *mt.ToSrvInteract, time.Duration) bool

var (
	digConds   = make(map[*Registerd[DigCond]]struct{})
	digCondsMu sync.RWMutex
)

// DigCond gets called before Place is acted upon
// If returns false doesn't place node (NodeDef.OnDig not called)
// Gets called BEFORE NodeDef.OnPlace
func RegisterDigCond(h DigCond) HookRef[Registerd[DigCond]] {
	placeCondsMu.Lock()
	defer placeCondsMu.Unlock()

	r := &Registerd[DigCond]{Caller(1), h}
	ref := HookRef[Registerd[DigCond]]{&digCondsMu, digConds, r}

	digConds[r] = struct{}{}

	return ref
}

type (
	NodeDigFunc         func(c *Client, i *mt.ToSrvInteract, digtime <-chan time.Duration)
	NodeStopDiggingFunc func(c *Client, i *mt.ToSrvInteract, digtime time.Duration)
	NodeDugFunc         func(c *Client, i *mt.ToSrvInteract, digtime time.Duration)
)

type NodeDef struct {
	mt.NodeDef

	OnDig         NodeDigFunc
	OnStopDigging NodeStopDiggingFunc
	OnDug         NodeDugFunc
}

var (
	nodeUnknown = &Registerd[NodeDef]{"builtin", NodeDef{NodeDef: mt.NodeDef{Param0: mt.Unknown, Name: "unknown"}}}
	nodeAir     = &Registerd[NodeDef]{"builtin", NodeDef{NodeDef: mt.NodeDef{Param0: mt.Air, Name: "air"}}}
	nodeIgnore  = &Registerd[NodeDef]{"builtin", NodeDef{NodeDef: mt.NodeDef{Param0: mt.Ignore, Name: "ignore"}}}

	nodeDefsMu sync.RWMutex
	nodeDefs   = map[string]*Registerd[NodeDef]{
		"unknown": nodeUnknown,
		"air":     nodeAir,
		"ignore":  nodeIgnore,
	}

	nodeDefsID = map[mt.Content]*Registerd[NodeDef]{
		mt.Unknown: nodeUnknown,
		mt.Air:     nodeAir,
		mt.Ignore:  nodeIgnore,
	}

	nodeDefID mt.Content // counts up for each node
)

// Add more item definitions to pool
// Param0 field is overwritten
func AddNodeDef(defs ...NodeDef) {
	nodeDefsMu.Lock()
	defer nodeDefsMu.Unlock()

	// add id
	for k := range defs {
		param0 := getNodeDefID()

		defs[k].Param0 = param0
		def := &Registerd[NodeDef]{Caller(1), defs[k]}

		nodeDefs[defs[k].Name] = def
		nodeDefsID[param0] = def
	}
}

func getNodeDefID() mt.Content {
	if nodeDefID == 125 {
		nodeDefID += 3
	}

	id := nodeDefID
	nodeDefID++

	return id
}

// RealNodeName returns the underlying name of a given node
// checks aliases map if !ok returns name
func RealNodeName(name string) string {
	aliasesMu.RLock()
	defer aliasesMu.RUnlock()

	rname, ok := aliases[name]
	if !ok {
		return name
	}

	return rname
}

// GetNodeDef returns pointer to node def if registerd
// otherwise nil
func GetNodeDef(name string) (def *Registerd[NodeDef]) {
	// check alias
	name = RealNodeName(name)

	nodeDefsMu.Lock()
	defer nodeDefsMu.Unlock()

	def, found := nodeDefs[name]
	if !found {
		return nil
	}

	return def
}

// GetNodeID returns pointer to node def if registerd
// otherwise nil
func GetNodeDefID(id mt.Content) (def *Registerd[NodeDef]) {
	nodeDefsMu.Lock()
	defer nodeDefsMu.Unlock()

	def, found := nodeDefsID[id]
	if !found {
		return nil
	}

	return def
}

// GetNodeID returns the Param0 of a node
// panics if not found
func GetNodeID(name string) mt.Content {
	def := GetNodeDef(name)

	if def == nil {
		Loggers.Warnf("%s\n", 1, nodeDefNotFoundErr{name})
		return mt.Unknown
	}

	return def.Thing.Param0
}

// NodeMaps generates a NodeIdMap and IdNodeMap
func NodeMaps() (NodeIdMap map[mt.Content]string, IdNodeMap map[string]mt.Content) {
	NodeIdMap = make(map[mt.Content]string)
	IdNodeMap = make(map[string]mt.Content)

	nodeDefsMu.RLock()
	defer nodeDefsMu.RUnlock()

	for k, v := range nodeDefsID {
		IdNodeMap[v.Thing.Name] = v.Thing.Param0
		NodeIdMap[k] = v.Thing.Name
	}

	return
}

// Send (cached) NodeDefinitions to client
func (c *Client) SendNodeDefs() (<-chan struct{}, error) {
	cmd := &mt.ToCltNodeDefs{
		Defs: nodeSlice(),
	}

	return c.SendCmd(cmd)
}

// nodeSlice generates a slice of all mt.NodeDefs
func nodeSlice() (s []mt.NodeDef) {
	nodeDefsMu.RLock()
	defer nodeDefsMu.RUnlock()

	for _, v := range nodeDefs {
		s = append(s, v.Thing.NodeDef)
	}

	return
}
