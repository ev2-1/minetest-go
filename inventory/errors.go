package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"errors"
	"fmt"
)

type ErrInvSetOutOfBounce struct {
	Index, Width uint32
}

func (e *ErrInvSetOutOfBounce) Error() string {
	return fmt.Sprintf("attempted access on inventory at index %d, but width is %d, this is probably a bug.", e.Index, e.Width)
}

type ErrInvLocationTypeNotHandled struct {
	Type string
}

func (e *ErrInvLocationTypeNotHandled) Error() string {
	return fmt.Sprintf("Inventory Location '%s' not handled", e.Type)
}

type ErrClientHasNoInventory struct {
	*minetest.Client
}

func (e *ErrClientHasNoInventory) Error() string {
	return fmt.Sprintf("client '%s' has no PlayerInv!", e.Name)
}

type ErrClientDosntHaveInventory struct {
	*minetest.Client
	Inv string
}

func (e *ErrClientDosntHaveInventory) Error() string {
	return fmt.Sprintf("client '%s' dosnt have Inventory '%s'!", e.Name, e.Inv)
}

type ErrPlayerInvNotExist struct {
	*minetest.Client
	Inv string
}

func (e *ErrPlayerInvNotExist) Error() string {
	return fmt.Sprintf("client '%s' dosn't have inventory '%s'", e.Name, e.Inv)
}

type ErrItemStackInsufficient struct {
	Note string

	InvName       string
	Stack, Action uint16
}

func (e *ErrItemStackInsufficient) Error() string {
	return fmt.Sprintf("stack in inv '%s' only is %d but tried to take %d (%s)", e.InvName, e.Stack, e.Action, e.Note)
}

type ErrInvNotAllowTake struct {
	Inv  string
	Slot uint32
}

func (e *ErrInvNotAllowTake) Error() string {
	return fmt.Sprintf("inventory '%s' dosn't allow item '%d' to be taken.", e.Inv, e.Slot)
}

type ErrInvNotAllowPut struct {
	Inv  string
	Slot uint32
}

func (e *ErrInvNotAllowPut) Error() string {
	return fmt.Sprintf("inventory '%s' dosn't allow item '%d' to be put.", e.Inv, e.Slot)
}

type ErrItemStackPresent struct {
	Inv  string
	Slot uint32

	Stack mt.Stack
}

func (e *ErrItemStackPresent) Error() string {
	return fmt.Sprintf("inventory '%s' already has item at %d (dst is: %d %s)", e.Inv, e.Slot, e.Stack.Count, e.Stack.Item.Name)
}

var (
	ErrInvUndefined = errors.New("Attempted action on undefined Inventory.")
)

type ErrPlayerInvNotPlayerInv struct {
	Name string
}

func (e *ErrPlayerInvNotPlayerInv) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("player inventory dosn't fullfil PlayerInv interface", e.Name)
	} else {
		return fmt.Sprintf("player inventory for player %s dosn't fullfil PlayerInv interface", e.Name)
	}
}
