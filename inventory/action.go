package inventory

import (
	"fmt"
	"io"
)

type InvAction interface {
	InvActionVerb() string

	String() string // String does NOT searialize
}

type InvActionMove struct {
	Count uint16

	From *InvLocation
	To   *InvLocation
}

func (act *InvActionMove) Deserialize(r io.Reader) {
	act.Count = ReadUint16(r)
	act.From = new(InvLocation)
	act.To = new(InvLocation)

	act.From.Deserialize(r)
	act.To.Deserialize(r)
}

func (act *InvActionMove) String() string {
	return fmt.Sprintf("Moving %d items from %s (inv: %s; stack: %d) to %s (inv: %s; stack: %d)",
		act.Count, act.From.Location, act.From.Name, act.From.Stack,
		act.To.Location, act.To.Name, act.To.Stack,
	)
}

func DeserializeInvAction(r io.Reader) (act InvAction, err error) {
	action := ReadString(r)

	newAction, ok := newInvAction[action]
	if !ok {
		return act, &ErrInvalidInvAction{action}
	}

	act = newAction()

	err = Deserialize(r, act)

	return
}

// ---
func (*InvActionMove) InvActionVerb() string { return "Move" }

var newInvAction = map[string]func() InvAction{
	"Move": func() InvAction { return new(InvActionMove) },
}
