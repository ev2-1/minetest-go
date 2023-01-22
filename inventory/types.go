package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"bytes"
	"io"
	"strconv"
	"strings"
	"sync"
)

type InvIdentifier interface {
	InvIdentifier() string
}

// InvIdentifierCurrentPlayer
func (*InvIdentifierCurrentPlayer) Deserialize(io.Reader) {}

type InvIdentifierCurrentPlayer struct{}

func (*InvIdentifierCurrentPlayer) InvIdentifier() string {
	return "current_player"
}

// InvIdentifierUndefined
type InvIdentifierUndefined struct{}

func (*InvIdentifierUndefined) InvIdentifier() string {
	return "undefined"
}

func (*InvIdentifierUndefined) Deserialize(io.Reader) {}

// InvIdentifierPlayer
type InvIdentifierPlayer struct {
	name string
}

func (*InvIdentifierPlayer) InvIdentifier() string {
	return "player"
}

func (i *InvIdentifierPlayer) Deserialize(r io.Reader) {
	i.name = ReadString(r, false)
}

// InvIdentifierNodeMeta
type InvIdentifierNodeMeta struct {
	X, Y, Z int16
}

func (*InvIdentifierNodeMeta) InvIdentifier() string {
	return "nodemeta"
}

func (i *InvIdentifierNodeMeta) Deserialize(r io.Reader) {
	cords := ReadString(r, false)
	vec := strings.SplitN(cords, ",", 4)
	if len(vec) != 3 {
		panic(SerializationError{ErrInvalidPos})
	}

	x, err := strconv.ParseInt(vec[0], 10, 16)
	if err != nil {
		panic(SerializationError{err})
	}

	y, err := strconv.ParseInt(vec[1], 10, 16)
	if err != nil {
		panic(SerializationError{err})
	}

	z, err := strconv.ParseInt(vec[2], 10, 16)
	if err != nil {
		panic(SerializationError{err})
	}

	i.X, i.Y, i.Z = int16(x), int16(y), int16(z)
}

// InvIdentifierDetached
type InvIdentifierDetached struct {
	name string
}

func (*InvIdentifierDetached) InvIdentifier() string {
	return "detached"
}

func (i *InvIdentifierDetached) Deserialize(r io.Reader) {
	i.name = ReadString(r, false)
}

var newInvIdentifier = map[string]func() InvIdentifier{
	"undefined":      func() InvIdentifier { return new(InvIdentifierUndefined) },
	"current_player": func() InvIdentifier { return new(InvIdentifierCurrentPlayer) },
	"player":         func() InvIdentifier { return new(InvIdentifierPlayer) },
	"nodemeta":       func() InvIdentifier { return new(InvIdentifierNodeMeta) },
	"detached":       func() InvIdentifier { return new(InvIdentifierDetached) },
}

// InvLocation
type InvLocation struct {
	Identifier InvIdentifier
	Name       string
	Stack      int
}

func (l *InvLocation) Deserialize(r io.Reader) {
	ident := ReadString(r, true)

	newId, ok := newInvIdentifier[ident]
	if !ok {
		newId = newInvIdentifier["undefined"]
	}

	newIdent := newId()
	newIdent.(Deserializer).Deserialize(r)

	l.Identifier = newIdent

	l.Name = ReadString(r, false)
	l.Stack = ReadInt(r, false)
}

func (l *InvLocation) Aquire(c *minetest.Client) (*RWInv, error) {
	switch indent := l.Identifier.(type) {
	case *InvIdentifierCurrentPlayer:
		return GetInv(c)

	default:
		_ = indent
		return nil, ErrInvalidLocation
	}
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
