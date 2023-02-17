package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"strings"
)

func map2slice[V comparable](m map[V]struct{}) []V {
	s := make([]V, len(m))

	var i int

	for k := range m {
		s[i] = k

		i++
	}

	return s
}

func strSlice[V any](s []V) []string {
	strs := make([]string, len(s))
	for k := range strs {
		strs[k] = fmt.Sprintf("%v", s[k])
	}

	return strs
}

type MultiError struct {
	Errs []error
}

func (err *MultiError) Error() string {
	return strings.Join(strSlice(err.Errs), " & ")
}

func (merr *MultiError) Add(err error) {
	merr.Errs = append(merr.Errs, err)
}

func Acks(acks ...<-chan struct{}) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		for _, ack := range acks {
			if ack == nil {
				continue
			}

			<-ack
		}

		close(ch)
	}()

	return ch
}

func IDAOMsgs(id mt.AOID, msgs ...mt.AOMsg) (s []mt.IDAOMsg) {
	s = make([]mt.IDAOMsg, len(msgs))

	for k := range msgs {
		s[k] = mt.IDAOMsg{
			ID:  id,
			Msg: msgs[k],
		}
	}

	return s
}

func PPos2AOPos(ppos minetest.PPos) AOPos {
	return AOPos{
		Pos: ppos.Pos.Pos,
		Rot: mt.Vec{1: ppos.Pos.Yaw},

		Dim: ppos.Dim,
	}
}
