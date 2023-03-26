package minetest

import (
	"github.com/anon55555/mt"

	"sync"
	"time"
)

const (
	LowestAOID  mt.AOID = (1)
	HighestAOID mt.AOID = (65534) // one lower than largest value
)

var (
	ActiveObjectsMu sync.RWMutex
	ActiveObjects   = make(map[mt.AOID]ActiveObject)
)

// getAOID returns next free aoid
func getAOID() mt.AOID {
	ActiveObjectsMu.RLock()
	defer ActiveObjectsMu.RUnlock()

	for id := LowestAOID; id < HighestAOID; id++ {
		if _, ok := ActiveObjects[id]; !ok {
			return id
		}
	}

	panic("No free AOIDs left!")
}

func registerAO(ao ActiveObject) {
	ActiveObjectsMu.Lock()
	defer ActiveObjectsMu.Unlock()

	ActiveObjects[ao.GetAO()] = ao
}

// ActiveObject
type ActiveObject interface {
	//SetAO should set the AOID
	SetAO(mt.AOID)

	//GetAO should return the ID
	//0 when none is defined
	GetAO() mt.AOID

	//AOInit should return Initialisation data for the AO
	AOInit(*Client) *AOInit

	Punch(clt *Client, i *mt.ToSrvInteract)

	//is called when removing AO
	Clean()
}

// ActiveObjectPlayer specifies a client ActiveObject
type ActiveObjectPlayer interface {
	ActiveObject

	//GetPos should return the position
	GetPos() PPos
	SetPos(PPos)
}

// ActiveObjectAPIAOPos specifies a standard interface to work with Positions of AOs
type ActiveObjectAPIAOPos interface {
	ActiveObject

	//GetPos should return the position
	GetAOPos() AOPos
	SetAOPos(AOPos)
}

// ActiveObjectTicker extends the ActiveObject to optional Ticked callbacks
type ActiveObjectTicker interface {
	ActiveObject

	// Tick gets called each PhysTick
	Tick(dtime time.Duration)
}

// ActiveObjectRelevant extends the ActiveObject to optionally overwrite the Relevance functions
type ActiveObjectRelevant interface {
	ActiveObject

	// Tick gets called when client is evaluating relevante of AO
	// true: relevant and will be added; false: won't
	Relevant(c *Client) bool
}

type AOPos struct {
	Pos [3]float32
	Rot [3]float32

	Dim DimID
}

func (aopos AOPos) AOPos() (pos mt.AOPos) {
	pos.Pos = aopos.Pos
	pos.Rot = aopos.Rot

	return pos
}

type AOInit struct {
	Name     string
	IsPlayer bool

	AOPos

	HP uint16

	AOMsgs []mt.AOMsg
}

func (i *AOInit) AOInitData(id mt.AOID) mt.AOInitData {
	return mt.AOInitData{
		Name:     i.Name,
		IsPlayer: i.IsPlayer,

		ID: id,

		Pos: i.Pos,
		Rot: i.Rot,

		Msgs: i.AOMsgs,
	}
}

func makeAOData() *AOData {
	return &AOData{
		AOs: make(map[mt.AOID]struct{}),

		Age: time.Now(),
	}
}

type AOData struct {
	sync.RWMutex

	Ready bool

	AOs map[mt.AOID]struct{}

	//is the clients AOID (client self does not have)
	AOID mt.AOID

	Age time.Time
}

func (cd *AOData) age() time.Duration {
	return time.Now().Sub(cd.Age)
}
