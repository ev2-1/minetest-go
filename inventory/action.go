package inventory

import (
	"github.com/anon55555/mt"
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
		c.Logf("[INV] %s", act.String())
	}

	var fromInv *RWInv

	// collect inventory
	fromInv, err = act.From.Aquire(c)
	if err != nil {
		return
	}

	fromInv.Lock()
	defer fromInv.Unlock()

	fromInvList := fromInv.M[act.From.Name]
	if len(fromInvList.Stacks) < act.From.Stack || fromInvList.Width < act.From.Stack {
		return
	}

	// ensure quantity
	if fromInvList.Stacks[act.From.Stack].Count < act.Count {
		return nil, ErrStackInsufficient
	}

	// Drop: TODO: make item actually drop though magic
	fromInvList.Stacks[act.From.Stack].Count -= act.Count

	fromInv.M[act.From.Name] = fromInvList

	// updating client:
	var str string
	str, err = fromInv.String()
	if err != nil {
		c.Logf("Error: %s", err)
		return
	}

	return c.SendCmd(
		&mt.ToCltInv{
			Inv: str,
		})
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
		c.Logf("[INV] %s", act.String())
	}

	var fromInv, toInv *RWInv

	// collect inventories
	fromInv, err = act.From.Aquire(c)
	if err != nil {
		return
	}

	fromInv.Lock()
	defer fromInv.Unlock()

	// only get a inv once, otherwise its gonna deadlock
	if act.From.Identifier == act.To.Identifier {
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
	var moveItem string

	fromInvList := fromInv.M[act.From.Name]
	if len(fromInvList.Stacks) < act.From.Stack || fromInvList.Width < act.From.Stack {
		return
	}

	moveItem = fromInvList.Stacks[act.From.Stack].Name

	// ensure quantity
	if fromInvList.Stacks[act.From.Stack].Count < act.Count {
		return nil, ErrStackInsufficient
	}

	// validate destination
	toInvList := toInv.M[act.To.Name]
	if len(toInvList.Stacks) < act.To.Stack || toInvList.Width < act.To.Stack {
		return nil, ErrInvalidStack
	}

	// check if slot is empty or same item:
	if !(toInvList.Stacks[act.To.Stack].Count == 0 || toInvList.Stacks[act.To.Stack].Name == moveItem) {
		return nil, ErrStackNotEmpty
	}

	// move:
	fromInvList.Stacks[act.From.Stack].Count -= act.Count
	toInvList.Stacks[act.To.Stack].Name = moveItem
	toInvList.Stacks[act.To.Stack].Count += act.Count

	toInv.M[act.To.Name] = toInvList
	fromInv.M[act.From.Name] = fromInvList

	// updating client:
	var str string
	str, err = fromInv.String()
	if err != nil {
		c.Logf("Error: %s", err)
		return
	}

	var ack1 <-chan struct{}

	// only send other inventory if inventories are different
	if fromInv != toInv {
		var str string
		str, err = fromInv.String()
		if err != nil {
			c.Logf("Error: %s", err)
			return
		}

		ack1, err = c.SendCmd(&mt.ToCltInv{
			Inv: str,
		})
	}

	ack2, err := c.SendCmd(&mt.ToCltInv{
		Inv: str,
	})

	ack := make(chan struct{})

	go func(a1, a2 <-chan struct{}, ack chan struct{}) {
		<-ack1
		<-ack2

		close(ack)
	}(ack1, ack2, ack)

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
