package inventory

import (
	"github.com/ev2-1/minetest-go/minetest"

	"fmt"
	"io"
)

type InvAction interface {
	InvActionVerb() string
	Apply(c *minetest.Client) (<-chan struct{}, error)

	String() string // String does NOT searialize
}

type InvActionDrop struct {
	Count uint16

	From *InvLocation
}

func (act *InvActionDrop) Deserialize(r io.Reader) {
	act.Count = ReadUint16(r, false)

	act.From = new(InvLocation)

	act.From.Deserialize(r)
}

func (act *InvActionDrop) String() string {
	return fmt.Sprintf("Dropping %d from %s (inv: %s; stack: %d)",
		act.Count,
		act.From.Identifier, act.From.Name, act.From.Stack,
	)
}

func (act *InvActionDrop) Apply(c *minetest.Client) (_ <-chan struct{}, err error) {
	if minetest.ConfigVerbose() {
		c.Logger.Printf("[INV] %s", act.String())
	}

	var fromInv RWInv

	// collect inventory
	fromInv, err = act.From.Aquire(c)
	if err != nil {
		return
	}

	fromInv.Lock()
	defer fromInv.Unlock()

	fromInvList, ok := fromInv.Get(act.From.Name)
	if !ok {
		return nil, ErrInvalidInv
	}

	// Ensure stack exists
	fromStack, ok := fromInvList.GetStack(act.From.Stack)
	if !ok {
		return nil, ErrInvalidStack
	}

	// ensure quantity
	if fromStack.Count < act.Count {
		return nil, ErrStackInsufficient
	}

	// Drop: TODO: make item actually drop though magic
	fromStack.Count -= act.Count
	fromInvList.SetStack(act.From.Stack, fromStack)

	fromInv.Set(act.From.Name, fromInvList)

	// updating client:
	var str string
	str, err = SerializeString(fromInv.Serialize)
	if err != nil {
		c.Logger.Printf("Error: %s", err)
		return
	}

	return act.From.SendUpdate(str, c)
}

type InvActionMove struct {
	Count uint16

	From *InvLocation
	To   *InvLocation
}

func (act *InvActionMove) Deserialize(r io.Reader) {
	act.Count = ReadUint16(r, false)

	act.From = new(InvLocation)
	act.To = new(InvLocation)

	act.From.Deserialize(r)
	act.To.Deserialize(r)
}

func (act *InvActionMove) String() string {
	return fmt.Sprintf("Moving %d items from %s (inv: %s; stack: %d) to %s (inv: %s; stack: %d)",
		act.Count, act.From.Identifier, act.From.Name, act.From.Stack,
		act.To.Identifier, act.To.Name, act.To.Stack,
	)
}

func (act *InvActionMove) Apply(c *minetest.Client) (_ <-chan struct{}, err error) {
	if minetest.ConfigVerbose() {
		c.Logger.Printf("[INV] %s", act.String())
	}

	var fromInv, toInv RWInv

	// collect inventories
	fromInv, err = act.From.Aquire(c)
	if err != nil {
		return
	}

	fromInv.Lock()
	defer fromInv.Unlock()

	// only get a inv once, otherwise its gonna deadlock
	if act.From.Identifier.Equals(act.To.Identifier) { // doesnt work for detached inv (atm)!
		toInv = fromInv
	} else {
		toInv, err = act.To.Aquire(c)
		if err != nil {
			return
		}

		toInv.Lock()
		defer toInv.Unlock()
	}

	// validate source:
	fromInvList, ok := fromInv.Get(act.From.Name)
	if !ok {
		return nil, ErrInvalidInv
	}

	fromStack, ok := fromInvList.GetStack(act.From.Stack)
	if !ok {
		return nil, ErrInvalidStack
	}

	moveItem := fromStack.Name

	// ensure quantity
	if fromStack.Count < act.Count {
		return nil, ErrStackInsufficient
	}

	// validate destination
	toInvList, ok := toInv.Get(act.To.Name)
	if !ok {
		return nil, ErrInvalidInv
	}

	if toInvList.Width() < act.To.Stack {
		return nil, ErrInvalidStack
	}

	toStack, ok := toInvList.GetStack(act.To.Stack)
	if !ok {
		return nil, ErrInvalidStack
	}

	// check if slot is empty or same item:
	if !(toStack.Count == 0 || toStack.Name == moveItem) {
		return nil, ErrStackNotEmpty
	}

	// move:
	//fromInvList.Stacks[act.From.Stack].Count -= act.Count
	fromStack.Count -= act.Count
	fromInvList.SetStack(act.From.Stack, fromStack)

	toStack.Name = moveItem
	toStack.Count += act.Count
	toInvList.SetStack(act.To.Stack, toStack)

	toInv.Set(act.To.Name, toInvList)
	fromInv.Set(act.From.Name, fromInvList)

	// updating client:
	fromStr, err := SerializeString(fromInv.Serialize)
	if err != nil {
		c.Logger.Printf("Error: %s", err)
		return
	}
	ack1, err := act.From.SendUpdate(fromStr, c)

	var ack2 <-chan struct{}

	// only send other inventory if inventories are different
	if fromInv != toInv {
		var toStr string
		toStr, err = SerializeString(toInv.Serialize)
		if err != nil {
			c.Logger.Printf("Error: %s", err)
			return
		}

		ack2, err = act.To.SendUpdate(toStr, c)
		if err != nil {
			return
		}
	}

	ack := make(chan struct{})

	go Acks(ack, ack1, ack2)

	return ack, err
}

func DeserializeInvAction(r io.Reader) (act InvAction, err error) {
	action := ReadString(r, false)

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
func (*InvActionDrop) InvActionVerb() string { return "Drop" }

var newInvAction = map[string]func() InvAction{
	"Move": func() InvAction { return new(InvActionMove) },
	"Drop": func() InvAction { return new(InvActionDrop) },
}
