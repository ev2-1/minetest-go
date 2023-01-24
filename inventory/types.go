package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"io"
	"strconv"
	"strings"
	"sync"
)

type InvIdentifier interface {
	InvIdentifier() string

	Equals(InvIdentifier) bool
}

// InvIdentifierUndefined
type InvIdentifierUndefined struct{}

func (*InvIdentifierUndefined) InvIdentifier() string {
	return "undefined"
}

func (*InvIdentifierUndefined) Equals(InvIdentifier) bool {
	return false // TODO
}

func (*InvIdentifierUndefined) Deserialize(io.Reader) {}

// InvIdentifierCurrentPlayer
func (*InvIdentifierCurrentPlayer) Deserialize(io.Reader) {}

type InvIdentifierCurrentPlayer struct{}

func (*InvIdentifierCurrentPlayer) Equals(i InvIdentifier) bool {
	_, ok := i.(*InvIdentifierCurrentPlayer)

	return ok
}

func (*InvIdentifierCurrentPlayer) InvIdentifier() string {
	return "current_player"
}

// InvIdentifierPlayer
type InvIdentifierPlayer struct {
	name string
}

func (*InvIdentifierPlayer) InvIdentifier() string {
	return "player"
}

func (*InvIdentifierPlayer) Equals(i InvIdentifier) bool {
	return false // TODO
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

func (*InvIdentifierNodeMeta) Equals(i InvIdentifier) bool {
	return false // TODO
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

func (d *InvIdentifierDetached) Equals(i InvIdentifier) bool {
	detached, ok := i.(*InvIdentifierDetached)
	if !ok {
		return ok
	}

	if detached.name == d.name {
		return true
	}

	return false
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

func (l *InvLocation) SendUpdate(list string, c *minetest.Client) (<-chan struct{}, error) {
	switch ident := l.Identifier.(type) {
	case *InvIdentifierCurrentPlayer:
		return c.SendCmd(&mt.ToCltInv{
			Inv: list,
		})

	case *InvIdentifierDetached:
		return c.SendCmd(&mt.ToCltDetachedInv{
			Name: ident.name,
			Keep: true,

			Inv: list,
		})

	default:
		_ = ident
	}

	return nil, ErrInvalidLocation
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

func (l *InvLocation) Aquire(c *minetest.Client) (RWInv, error) {
	switch indent := l.Identifier.(type) {
	case *InvIdentifierCurrentPlayer:
		return GetInv(c)

	case *InvIdentifierDetached:
		return GetDetached(indent.name, c)

	default:
		_ = indent
		return nil, ErrInvalidLocation
	}
}

type RWInv interface {
	Inv

	RLock()
	RUnlock()
	Lock()
	Unlock()
}

type Inv interface {
	Get(string) (InvList, bool)
	Set(string, InvList)

	Serialize(io.Writer) error
}

type InvList interface {
	Width() int
	GetStack(int) (mt.Stack, bool)
	SetStack(int, mt.Stack) bool

	Serialize(io.Writer) error

	InvList() mt.InvList
}

type SimpleInv struct {
	M map[string]InvList

	sync.RWMutex
}

func (si *SimpleInv) Get(k string) (l InvList, ok bool) {
	l, ok = si.M[k]

	return
}

func (si *SimpleInv) Set(k string, v InvList) {
	si.M[k] = v
}

func (si *SimpleInv) Serialize(w io.Writer) error {
	return si.Inv().Serialize(w)
}

func (inv *SimpleInv) Deserialize(w io.Reader) (err error) {
	list := new(mt.Inv)
	if err = list.Deserialize(w); err != nil {
		return
	}

	SimpleInvFromNamedInvList(*list, inv)

	return
}

type SimpleInvList struct {
	List mt.InvList
}

func (il *SimpleInvList) Width() int {
	if len(il.List.Stacks) > il.List.Width {
		return len(il.List.Stacks)
	} else {
		return il.List.Width
	}
}

func (il *SimpleInvList) GetStack(i int) (s mt.Stack, ok bool) {
	if len(il.List.Stacks) < i {
		ok = false

		return
	} else {
		return il.List.Stacks[i], true
	}
}

func (il *SimpleInvList) SetStack(i int, s mt.Stack) bool {
	if len(il.List.Stacks) < i {
		return false
	} else {
		il.List.Stacks[i] = s
		return true
	}
}

func (il *SimpleInvList) Serialize(w io.Writer) error {
	return il.List.Serialize(w)
}

func (il *SimpleInvList) InvList() mt.InvList {
	return il.List
}

func (inv *SimpleInv) Inv() (r mt.Inv) {
	for k, v := range inv.M {
		r = append(r, mt.NamedInvList{
			Name:    k,
			InvList: v.InvList(),
		})
	}

	return
}

//func (inv *Inv) Set(name string, list mt.NamedInvList) {}

// fulfills the ClientDataSerialize Interface
//func (inv *Inv) Serialize(w io.Writer) (err error) {
//	return inv.Inv().Serialize(w)
//}

func SimpleInvFromNamedInvList(list mt.Inv, inv *SimpleInv) {
	inv.M = make(map[string]InvList)

	for i := 0; i < len(list); i++ {
		inv.M[list[0].Name] = &SimpleInvList{list[0].InvList}
	}

	return
}
