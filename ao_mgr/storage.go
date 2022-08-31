package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
	"sync"
)

var activeObjects = make(map[mt.AOID]ActiveObject)
var activeObjectsMu sync.RWMutex

func GetAO(id mt.AOID) ActiveObject {
	activeObjectsMu.RLock()
	defer activeObjectsMu.RUnlock()

	return activeObjects[id]
}

func GetAOPos(id mt.AOID) (bool, mt.AOPos) {
	activeObjectsMu.RLock()
	defer activeObjectsMu.RUnlock()

	p, f := activeObjects[id]
	if !f {
		return f, mt.AOPos{}
	}

	return f, p.GetPos()
}

// ActiveObject describes a active object.
// an example struct is located in example.go (`ActiveObjectS`)
type ActiveObject interface {
	SetID(mt.AOID)
	GetID() mt.AOID

	// Pkts retruns all changed aofields as msgs and a boolean if any changed
	Pkts() ([]mt.AOMsg, bool)

	// DoPhysics gets called at most once per tick with dtime to do physics
	DoPhysics(dtime float32)

	GetPos() mt.AOPos
	SetPos(mt.AOPos)

	GetBonePos(string) (mt.AOBonePos, bool)
	SetBonePos(string, mt.AOBonePos)

	GetBones() map[string]mt.AOBonePos

	GetProps() mt.AOProps
	GetArmor() []mt.Group

	InitPkt(*minetest.Client) mt.AOInitData

	Interact(AOInteract)

	Delete(DelReason)
}

type DelReason uint8

const (
	ClearObjects DelReason = iota
	ForceKill
)

// AOInteract describes a interaction with a active object
type AOInteract struct {
	Player *minetest.Client

	Action   mt.Interaction
	ItemSlot uint16
	Pos      mt.PlayerPos

	Anim mt.AOAnim
}

type AOPhys struct {
	Walk, Jump, Gravity float32
}

func (p AOPhys) AOPhysOverride() mt.AOPhysOverride {
	return mt.AOPhysOverride{
		Walk:    p.Walk,
		Jump:    p.Jump,
		Gravity: p.Gravity,
	}
}

// returns true if Walk|Jump|Gravety is not 1
func (p AOPhys) NotDefault() bool {
	return p.Walk != 1 || p.Jump != 1 || p.Gravity != 1
}

// pkt returns a mt.ToCltAOPhysOverride
func (p AOPhys) Pkt() *mt.AOCmdPhysOverride {
	return &mt.AOCmdPhysOverride{
		Phys: mt.AOPhysOverride{
			Walk:    p.Walk,
			Jump:    p.Jump,
			Gravity: p.Gravity,
		},
	}
}

type AOState struct {
	Anim mt.AOAnim

	PosMu sync.RWMutex
	Pos   mt.AOPos

	Armor  []mt.Group
	Attach mt.AOAttach

	Phys AOPhys

	BonesMu sync.RWMutex
	Bones   map[string]mt.AOBonePos

	HP uint16

	TextureMod mt.Texture
}

func (s *AOState) GetPhys() AOPhys {
	return s.Phys
}

func (s *AOState) GetAnimState() mt.AOAnim {
	return s.Anim
}

func (s *AOState) GetArmor() []mt.Group {
	return s.Armor
}

func (s *AOState) GetAttach() mt.AOAttach {
	return s.Attach
}

func (s *AOState) GetBones() map[string]mt.AOBonePos {
	m := make(map[string]mt.AOBonePos)

	s.BonesMu.RLock()
	defer s.BonesMu.RUnlock()

	for k, v := range s.Bones {
		m[k] = v
	}

	return m
}

func (s *AOState) GetHP() uint16 {
	return s.HP
}

func (s *AOState) GetTextureMod() mt.Texture {
	return s.TextureMod
}

func (s *AOState) GetPos() mt.AOPos {
	s.PosMu.RLock()
	defer s.PosMu.RUnlock()

	return s.Pos
}

func (s *AOState) SetPos(p mt.AOPos) {
	s.PosMu.Lock()
	defer s.PosMu.Unlock()

	s.Pos = p
}
