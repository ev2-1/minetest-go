package minetest

const RelevantDistance float32 = 100 * 10 // in 10th nodes

func Relevant(ao ActiveObject, clt *Client) bool {
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
