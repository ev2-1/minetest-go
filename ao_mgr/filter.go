package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

var (
	ReleventDistance float32 = 1000 // in nodes/10; distance around player their still informed about AOs
)

func RelevantAO(clt *minetest.Client, ao ActiveObject) bool {
	// check if ID is the one of client
	cpos := minetest.GetPos(clt)
	aopos := ao.GetPos()

	dis := Distance(mt.Vec(aopos.Pos.Pos), cpos.Pos)
	if cpos.Dim == aopos.Dim && dis < ReleventDistance {
		return true
	}

	return false
}

/*func FilterRelevantRms(clt *minetest.Client, rms []mt.AOID) (r []mt.AOID) {
	for _, rm := range rms {
		f, p := GetAOPos(rm)
		if !f {
			break
		}

		// check if is loaded by client
	}

	return
}

func FilterRelevantMsgs(pos mt.Pos, msgs []mt.IDAOMsg) (r []mt.IDAOMsg) {
	/*	for _, msg := range msgs {
			f, p := GetAOPos(msg.ID)
			if !f {
				break
			}

			if Distance(mt.Vec(pos), mt.Vec(p.Pos)) > ReleventDistance {
				r = append(r, msg)
			}
		}
*/ /*

	return msgs
}
*/
