package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

// ActiveObjectS is a example for a ActiveObject interface
type ActiveObjectS struct {
	AOState

	ID mt.AOID

	AnimSpeed float32
	Attach    mt.AOAttach
	Props     mt.AOProps
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

func (ao *ActiveObjectS) SetID(id mt.AOID) {
	ao.ID = id
}

func (ao *ActiveObjectS) GetID() mt.AOID {
	return ao.ID
}

func (ao *ActiveObjectS) GetBonePos(str string) (p mt.AOBonePos, ok bool) {
	p, ok = ao.Bones[str]
	return
}

func (ao *ActiveObjectS) GetProps() mt.AOProps {
	return ao.Props
}

func (ao *ActiveObjectS) Delete(DelReason) {
}

func (ao *ActiveObjectS) Interact(AOInteract) {
}

func (ao *ActiveObjectS) InitPkt(id mt.AOID, clt *minetest.Client) mt.AOInitData {
	var msgs []mt.AOMsg

	appendM := func(msgss ...mt.AOMsg) {
		msgs = append(msgs, msgss...)
	}

	appendM(&mt.AOCmdProps{
		Props: ao.GetProps(),
	})

	anim := ao.GetAnimState()
	if anim.Active {
		appendM(&mt.AOCmdAnim{
			Anim: anim.AOAnim,
		})
	}

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

	return mt.AOInitData{
		IsPlayer: false,

		ID:  id,
		Pos: ao.GetPos().Pos,
		Rot: ao.GetPos().Rot,

		HP: ao.GetHP(),

		Msgs: msgs,
	}
}
