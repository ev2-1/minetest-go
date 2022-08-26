package ao

import (
	"github.com/anon55555/mt"
	//	"github.com/ev2-1/minetest-go-plugins/tools/pos"

	"github.com/g3n/engine/math32"
)

var (
	ReleventDistance float32 = 100 // in nodes; distance around player their still informed about AOs
)

func FilterRelevantAdds(pos mt.Pos, adds []mt.AOAdd) (r []mt.AOAdd) {
	// mt.AOAdd.InitData.Pos (mt.Pos = mt.Vec)
	for _, add := range adds {
		if Distance(mt.Vec(pos), mt.Vec(add.InitData.Pos)) > ReleventDistance {
			r = append(r, add)
		}
	}

	return
}

func FilterRelevantRms(pos mt.Pos, rms []mt.AOID) (r []mt.AOID) {
	for _, rm := range rms {
		f, p := GetAOPos(rm)
		if !f {
			break
		}

		if Distance(mt.Vec(pos), mt.Vec(p.Pos)) > ReleventDistance {
			r = append(r, rm)
		}
	}

	return
}

func FilterRelevantMsgs(pos mt.Pos, msgs []mt.IDAOMsg) (r []mt.IDAOMsg) {
	for _, msg := range msgs {
		f, p := GetAOPos(msg.ID)
		if !f {
			break
		}

		if Distance(mt.Vec(pos), mt.Vec(p.Pos)) > ReleventDistance {
			r = append(r, msg)
		}
	}

	return
}

func Distance(a, b mt.Vec) float32 {
	var number float32

	number += math32.Pow((a[0] - b[0]), 2)
	number += math32.Pow((a[1] - b[1]), 2)
	number += math32.Pow((a[2] - b[2]), 2)

	return math32.Sqrt(number)
}
