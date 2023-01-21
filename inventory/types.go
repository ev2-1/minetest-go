package inventory

import (
	"github.com/anon55555/mt"

	"bytes"
	"io"
	"sync"
)

type InvLocation struct {
	Location string
	Name     string
	Stack    int
}

func (l *InvLocation) Deserialize(r io.Reader) {
	l.Location = ReadString(r)
	l.Name = ReadString(r)
	l.Stack = ReadInt(r)
}

type RWInv struct {
	sync.RWMutex
	*Inv
}

type Inv struct {
	M map[string]*mt.InvList
}

func (inv *Inv) Inv() (r mt.Inv) {
	for k, v := range inv.M {
		r = append(r, mt.NamedInvList{
			Name:    k,
			InvList: *v,
		})
	}

	return
}

func (inv *Inv) String() (s string, err error) {
	sbuf := &bytes.Buffer{}
	if err = inv.Serialize(sbuf); err != nil {
		return
	}

	s = sbuf.String()

	return
}

func (inv *Inv) Set(name string, list mt.NamedInvList) {}

// fulfills the ClientDataSerialize Interface
func (inv *Inv) Serialize(w io.Writer) (err error) {
	return inv.Inv().Serialize(w)
}

func (inv *RWInv) Deserialize(w io.Reader) (err error) {
	inv.Inv = new(Inv)

	return inv.Inv.Deserialize(w)
}

func (inv *Inv) Deserialize(w io.Reader) (err error) {
	list := new(mt.Inv)
	if err = list.Deserialize(w); err != nil {
		return
	}

	InvFromNamedInvList(*list, inv)

	return
}

func InvFromNamedInvList(list mt.Inv, inv *Inv) {
	inv.M = make(map[string]*mt.InvList)

	for i := 0; i < len(list); i++ {
		inv.M[list[0].Name] = &list[0].InvList
	}

	return
}
