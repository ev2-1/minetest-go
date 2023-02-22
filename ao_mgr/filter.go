package ao

import (
	"github.com/ev2-1/minetest-go/minetest"

	"github.com/g3n/engine/math32"
)

const RelevantDistance float32 = 100 * 10 // in 10th nodes

func Relevant(ao ActiveObject, clt *minetest.Client) bool {
	if relao, ok := ao.(ActiveObjectRelevant); ok {
		return relao.Relevant(clt)
	}

	// Default Relevance function:
	if posao, ok := ao.(ActiveObjectAPIAOPos); ok {
		aopos := posao.GetAOPos()
		cltpos := clt.GetPos()

		return aopos.Dim == cltpos.Dim &&
			Distance(aopos.Pos, cltpos.Pos.Pos) <= RelevantDistance
	}

	// default true:
	return true
}

func Distance(a, b [3]float32) float32 {
	var number float32

	number += math32.Pow((a[0] - b[0]), 2)
	number += math32.Pow((a[1] - b[1]), 2)
	number += math32.Pow((a[2] - b[2]), 2)

	return math32.Sqrt(number)
}
