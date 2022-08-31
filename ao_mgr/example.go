package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"sync"
)

type AOField uint8

const (
	FieldPos AOField = iota
	FieldAnimState
	FieldArmor
	FieldAttach
	FieldPhys
	FieldBonePos
	FieldHP
	FieldTextureMod

	FieldProps
)

// ActiveObjectS is a example for a ActiveObject interface
type ActiveObjectS struct {
	AOState

	ID mt.AOID

	Props mt.AOProps

	BonePosChanged   []string
	BonePosChangedMu sync.RWMutex

	ChangedMu sync.RWMutex
	Changed   map[AOField]struct{}
}

/*
	SetID(mt.AOID)
	GetID() mt.AOID

	GetPos() mt.AOPos
	GetBonePos(string) mt.AOBonePos

	GetProps() mt.AOProps
	GetArmor() []mt.Group

	Intet
*/

func (ao *ActiveObjectS) changed(f AOField) {
	ao.ChangedMu.Lock()
	defer ao.ChangedMu.Unlock()

	if ao.Changed == nil {
		ao.Changed = make(map[AOField]struct{})
	}

	ao.Changed[f] = struct{}{}
}

func (ao *ActiveObjectS) SetPos(p mt.AOPos) {
	ao.AOState.SetPos(p)

	ao.changed(FieldPos)
}

func (ao *ActiveObjectS) Pkts() (m []mt.AOMsg, b bool) {
	ao.ChangedMu.Lock()
	defer ao.ChangedMu.Unlock()

	if ao.Changed == nil {
		return
	}

	i := 0
	m = make([]mt.AOMsg, len(ao.Changed))

	for c, _ := range ao.Changed {
		switch c {
		case FieldPos:
			m[i] = &mt.AOCmdPos{
				Pos: ao.GetPos(),
			}
			break

		case FieldAnimState:
			m[i] = &mt.AOCmdAnim{
				Anim: ao.Anim,
			}
			break

		case FieldArmor:
			m[i] = &mt.AOCmdArmorGroups{
				Armor: ao.Armor,
			}
			break

		case FieldAttach:
			m[i] = &mt.AOCmdAttach{
				Attach: ao.Attach,
			}
			break

		case FieldPhys:
			m[i] = &mt.AOCmdPhysOverride{
				Phys: ao.Phys.AOPhysOverride(),
			}
			break

		case FieldBonePos:
			ao.BonePosChangedMu.Lock()
			if len(ao.BonePosChanged) == 0 {
				break
			}

			bone := ao.BonePosChanged[0]
			pos, ok := ao.GetBonePos(bone)
			if !ok {
				break
			}

			m[i] = &mt.AOCmdBonePos{
				Bone: bone,
				Pos:  pos,
			}

			ao.BonePosChanged = ao.BonePosChanged[1:]

			if len(ao.BonePosChanged) != 0 {
				for _, bone := range ao.BonePosChanged {
					pos, ok := ao.GetBonePos(bone)
					if ok {
						m = append(m, &mt.AOCmdBonePos{
							Bone: bone,
							Pos:  pos,
						})
					}
				}
			}

			ao.BonePosChanged = nil
			ao.BonePosChangedMu.Unlock()
			break

		case FieldHP:
			m[i] = &mt.AOCmdHP{
				HP: ao.HP,
			}
			break

		case FieldTextureMod:
			m[i] = &mt.AOCmdTextureMod{
				Mod: ao.TextureMod,
			}
			break

		case FieldProps:
			m[i] = &mt.AOCmdProps{
				Props: ao.Props,
			}
			break
		}

		i++
	}

	ao.Changed = nil

	return m, true
}

func (ao *ActiveObjectS) SetID(id mt.AOID) {
	ao.ID = id
}

func (ao *ActiveObjectS) GetID() mt.AOID {
	return ao.ID
}

func (ao *ActiveObjectS) GetBonePos(str string) (p mt.AOBonePos, ok bool) {
	ao.BonesMu.RLock()
	defer ao.BonesMu.RUnlock()

	p, ok = ao.Bones[str]
	return
}

func (ao *ActiveObjectS) SetBonePos(str string, p mt.AOBonePos) {
	ao.BonesMu.Lock()
	ao.BonePosChangedMu.Lock()
	defer ao.BonesMu.Unlock()
	defer ao.BonePosChangedMu.Unlock()

	if _, ok := ao.Bones[str]; ok {
		ao.Bones[str] = p
	}

	ao.BonePosChanged = append(ao.BonePosChanged, str)
	ao.changed(FieldBonePos)
}

func (ao *ActiveObjectS) GetProps() mt.AOProps {
	return ao.Props
}

func (ao *ActiveObjectS) Delete(DelReason) {
}

func (ao *ActiveObjectS) Interact(AOInteract) {
}

func (ao *ActiveObjectS) InitPkt(clt *minetest.Client) mt.AOInitData {
	var msgs []mt.AOMsg

	appendM := func(msgss ...mt.AOMsg) {
		msgs = append(msgs, msgss...)
	}

	appendM(&mt.AOCmdProps{
		Props: ao.GetProps(),
	})

	appendM(&mt.AOCmdAnim{
		Anim: ao.GetAnimState(),
	})

	// bones:
	bones := ao.GetBones()

	if len(bones) != 0 {
		for name, pos := range bones {
			appendM(&mt.AOCmdBonePos{
				Bone: name,
				Pos:  pos,
			})
		}
	}

	// phys
	phys := ao.GetPhys()
	if phys.NotDefault() {
		appendM(phys.Pkt())
	}

	// finnal finish AO
	appendM(&mt.AOCmdAttach{
		Attach: mt.AOAttach{ParentID: 0, ForceVisible: true},
	})

	p := ao.GetPos()

	return mt.AOInitData{
		IsPlayer: false,

		ID:  ao.GetID(),
		Pos: p.Pos,
		Rot: p.Rot,

		HP: ao.GetHP(),

		Msgs: msgs,
	}
}
