package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
)

type InvDef map[string]int

type PlayerInv struct {
	sync.RWMutex

	Formspec string
	Inv      map[string]Inv
}

func (inv *PlayerInv) Serialize(w io.Writer) {
	inv.RLock()
	defer inv.RUnlock()

	for n, l := range inv.Inv {
		fmt.Fprintln(w, "List", n, l.Width())
		l.Serialize(w)
	}
	fmt.Fprintln(w, "EndInventory")
}

type Inv interface {
	Width() uint32
	Get(slot uint32) mt.Stack
	Set(slot uint32, stack mt.Stack) error

	AllowTake(c *minetest.Client, slot uint32) bool
	AllowPut(c *minetest.Client, slot uint32) bool

	RLock()
	RUnlock()
	Lock()
	Unlock()

	Copy() mt.InvList
	Serialize(w io.Writer) error
	SerializeKeep(w io.Writer, oldinv mt.InvList) error
	Type() InvLocationType
}

type PlayerInvI interface {
	Inv

	PlayerName() string
}

// see https://github.com/minetest/minetest/blob/2d8eac4e0a609acf7a26e59141e6c684fdb546d0/src/inventorymanager.cpp#L45
type InvLocationType uint8

//go:generate stringer -type=InvLocationType -linecomment
const (
	InvLocationUndefined     InvLocationType = iota // undefined
	InvLocationCurrentPlayer                        // current_player
	InvLocationPlayer                               // player
	InvLocationNodeMeta                             // nodemeta
	InvLocationDetached                             // Srcdetached
)

func (t InvLocationType) SendKeep(c *minetest.Client, inv Inv, keep mt.InvList) (ack <-chan struct{}, err error) {
	buf := &bytes.Buffer{}

	switch t {
	case InvLocationUndefined:
		log.Printf("Warning, tried to send InvLocationUndefined to clt %s \n", c.Name)
		return nil, ErrInvUndefined
		break

	case InvLocationCurrentPlayer:
		err := inv.SerializeKeep(buf, keep)
		if err != nil {
			return nil, err
		}

		return c.SendCmd(&mt.ToCltInv{
			Inv: buf.String(),
		})

	case InvLocationPlayer:
		pi, ok := inv.(PlayerInvI)
		if !ok {
			err := &ErrPlayerInvNotPlayerInv{}
			log.Printf("ERROR, Attemtempted SendKeep InvLocationPlayer: %s \n", err)
		}

		err := inv.SerializeKeep(buf, keep)
		if err != nil {
			return nil, err
		}

		return c.SendCmd(&mt.ToCltDetachedInv{
			Name: "player:" + pi.PlayerName(), // TODO
			Keep: true,
			// Len: 0,

			Inv: buf.String(),
		})

	case InvLocationNodeMeta:
	case InvLocationDetached:
	}

	return nil, ErrInvUndefined
}

func (t InvLocationType) Deserialize(str string) (InvLocationType, error) {
	switch str {
	case "undefined":
		return InvLocationUndefined, nil
	case "current_player":
		return InvLocationCurrentPlayer, nil
	case "player":
		return InvLocationPlayer, nil
	case "nodemeta":
		return InvLocationNodeMeta, nil
	case "detached":
		return InvLocationDetached, nil
	}

	return t, &ErrInvLocationTypeNotHandled{Type: str}
}

type InvLocation struct {
	Type InvLocationType

	Somewhere bool

	Inventory string
	Slot      uint32

	Name string   // only present with `InvLocationPlayer` and `InvLocationDetached`
	P    [3]int16 // only present with `InvLocationDetached`
}

func (l InvLocation) Serialize() string {
	var locationString string

	switch l.Type {
	case InvLocationCurrentPlayer:
		locationString = "current_player"
	case InvLocationPlayer:
		locationString = fmt.Sprintf("player:%s", l.Name)
	case InvLocationNodeMeta:
		locationString = fmt.Sprintf("nodemeta:%d,%d,%d", l.P[0], l.P[1], l.P[2])
	case InvLocationDetached:
		locationString = fmt.Sprintf("detached:%s", l.Name)

	case InvLocationUndefined:
	default:
		locationString = "undefined"
	}

	if l.Somewhere {
		return fmt.Sprintf("%s %s", locationString, l.Inventory)
	} else {
		return fmt.Sprintf("%s %s %d", locationString, l.Inventory, l.Slot)
	}
}

func (l *InvLocation) Deserialize(r *bufio.Reader, somewhere bool) (*InvLocation, error) {
	var err error

	locationString, err := r.ReadString(' ')
	if err != nil {
		return l, err
	}

	locationString = strings.TrimSuffix(locationString, " ")

	t := strings.SplitN(locationString, ":", 2)[0]
	l.Type, err = l.Type.Deserialize(t)
	if err != nil {
		return l, err
	}

	l.Inventory, err = r.ReadString(' ')
	if err != nil {
		return l, err
	}

	l.Inventory = strings.TrimSuffix(l.Inventory, " ")

	l.Somewhere = somewhere

	if !somewhere {
		slot, err := r.ReadString(' ')
		if err != nil && !errors.Is(err, io.EOF) {
			return l, err
		}

		slot = strings.TrimSuffix(slot, " ")

		i, err := strconv.Atoi(slot)
		if err != nil {
			return l, err
		}

		l.Slot = uint32(i)
	}

	return l, nil
}

type Action interface {
	GetCount() uint16
	Serialize() string

	ActionString() string

	Apply(c *minetest.Client) error
}

type ActionMove struct {
	Src InvLocation
	Dst InvLocation

	Count uint16
}

func (a *ActionMove) GetCount() uint16 {
	return a.Count
}

func (a *ActionMove) Serialize() string {
	return fmt.Sprintf("Move %d %s %s", a.Count, a.Src.Serialize(), a.Dst.Serialize())
}

func (a *ActionMove) ActionString() string {
	return "Move"
}

func DeserializeActionMove(r *bufio.Reader, somewhere bool) (a *ActionMove, err error) {
	var str string
	a = &ActionMove{}

	str, err = r.ReadString(' ')
	if err != nil {
		return
	}

	c, err := strconv.Atoi(strings.TrimSuffix(str, " "))
	if err != nil {
		return
	}

	a.Count = uint16(c)

	_, err = a.Src.Deserialize(r, false)
	if err != nil {
		return

	}

	_, err = a.Dst.Deserialize(r, somewhere)
	if err != nil {
		return
	}

	return
}

func (a *ActionMove) deserialize(r *bufio.Reader, somewhere bool) (err error) {
	a, err = DeserializeActionMove(r, somewhere)

	return
}

func (a *ActionMove) Apply(c *minetest.Client) error {
	var currentPlayerInv *PlayerInv

	if a.Src.Type == InvLocationUndefined || a.Dst.Type == InvLocationUndefined {
		return ErrInvUndefined
	}

	// Lock all inventories that are about to be changed:
	if a.Src.Type == InvLocationCurrentPlayer || a.Dst.Type == InvLocationCurrentPlayer {
		currentPlayerInv = GetPlayerInv(c)
		if currentPlayerInv == nil {
			return &ErrClientHasNoInventory{c}
		}

		currentPlayerInv.Lock()
		defer currentPlayerInv.Unlock()
	}

	var srcItem mt.Item
	var src, dst Inv

	// Verify src sufficient:
	switch a.Src.Type {
	case InvLocationCurrentPlayer:
		inv, ok := currentPlayerInv.Inv[a.Src.Inventory]
		if !ok {
			return &ErrClientDosntHaveInventory{c, a.Src.Inventory}
		}

		inv.Lock()
		defer inv.Unlock()

		if !inv.AllowTake(c, a.Src.Slot) {
			return &ErrInvNotAllowTake{a.Src.Inventory, a.Src.Slot}
		}

		stack := inv.Get(a.Src.Slot)
		if stack.Count < a.Count {
			return &ErrItemStackInsufficient{"player " + c.Name, a.Src.Inventory, stack.Count, a.Count}
		}

		src = inv
		srcItem = stack.Item
	}

	// Verify dst
	switch a.Dst.Type {
	case InvLocationCurrentPlayer:
		inv, ok := currentPlayerInv.Inv[a.Dst.Inventory]
		if !ok {
			return &ErrPlayerInvNotExist{c, a.Dst.Inventory}
		}

		// only lock again if two different inventories
		if src != inv {
			inv.Lock()
			defer inv.Unlock()
		}

		if !inv.AllowPut(c, a.Dst.Slot) {
			return &ErrInvNotAllowPut{a.Dst.Inventory, a.Dst.Slot}
		}

		stack := inv.Get(a.Dst.Slot)
		if stack.Count != 0 && stack.Item != srcItem { // TODO stack size!
			return &ErrItemStackPresent{a.Dst.Inventory, a.Dst.Slot, stack}
		}

		dst = inv
	}

	// store old values, for diff later
	oldSrc, oldDst := src.Copy(), dst.Copy()

	// Apply:
	srcStack := src.Get(a.Src.Slot)
	srcStack.Count -= a.Count

	dstStack := dst.Get(a.Dst.Slot)
	dstStack.Item = srcItem
	dstStack.Count += a.Count

	src.Set(a.Src.Slot, srcStack)
	dst.Set(a.Dst.Slot, dstStack)

	updateInv <- src

	// send client updated inv(s):
	// update src:
	src.Type().SendKeep(c, src, oldSrc)

	// if dst isn't same, send as well
	if src != dst {
		dst.Type().SendKeep(c, dst, oldDst)
	}

	/*
		InvLocationUndefined     InvLocationType = iota // undefined
		InvLocationCurrentPlayer                        // current_player
		InvLocationPlayer                               // player
		InvLocationNodeMeta                             // nodemeta
		InvLocationDetached
	*/

	/*
		TODO: implement
		"player" // how to authorize? req: mby perms?
		"nodemeta" // anticheat! req: configuration
		"detached" // anticheat! does exist for client?
	*/

	return nil
}

var (
	ErrActionNotHandled = errors.New("Action Not handled.")
)

// Deserialize Action
func DeserializeAction(r *bufio.Reader) (Action, error) {
	action, err := r.ReadString(' ')
	if err != nil {
		return nil, err
	}

	switch strings.TrimSuffix(action, " ") {
	case "Move":
		a, err := DeserializeActionMove(r, false)
		if err != nil {
			return nil, err
		}

		return a, nil

	case "MoveSomewhere":
		a, err := DeserializeActionMove(r, true)
		if err != nil {
			return nil, err
		}

		return a, nil
	case "Drop":
	case "Craft":
	}

	return nil, ErrActionNotHandled
}
