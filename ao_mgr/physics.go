package ao

import (
	"github.com/anon55555/mt"
)

func (ao *ActiveObjectS) DoPhysics(dtime float32) {
	pos := ao.GetPos()

	oldVel := pos.Vel
	pos.Vel = addVec(mulVec(pos.Acc, dtime), pos.Vel)
	pos.Pos = mt.Pos(addVec(
		mt.Vec(pos.Pos),
		mulVecs(
			mt.Vec(pos.Pos),
			mulVec(
				addVec(
					oldVel,
					pos.Vel,
				),
				0.5,
			),
		),
	),
	)

	ao.SetPosPhys(pos)
	//ao.SetPos(pos)
}

/*
	old_speed = speed
	speed += dtime * acceleration
	pos += dtime * (old_speed + speed) / 2
*/
