package inventory

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"io"
	"sync"
)

// StorageInv is a basic inventory for just storage
// no active componentes
// Can't be used as PlayerInvI without additional code
type StorageInv struct {
	mt.InvList // aka. inventory with name

	sync.RWMutex
}

// Use only for Keep, esc functions, not to change!
func (inv *StorageInv) Copy() (il mt.InvList) {
	copy(il.Stacks, inv.InvList.Stacks)
	il.Width = inv.InvList.Width

	return
}

func (inv *StorageInv) Width() uint32 {
	return uint32(inv.InvList.Width)
}

func (inv *StorageInv) Get(slot uint32) mt.Stack {
	if inv.Width() < slot {
		return mt.Stack{}
	}

	return inv.InvList.Stacks[slot]
}

func (inv *StorageInv) Set(slot uint32, stack mt.Stack) error {
	if inv.Width() < slot {
		return &ErrInvSetOutOfBounce{slot, inv.Width()}
	}

	inv.InvList.Stacks[slot] = stack

	return nil
}

func (inv *StorageInv) Serialize(w io.Writer) error {
	return inv.InvList.Serialize(w)
}

func (inv *StorageInv) SerializeKeep(w io.Writer, old mt.InvList) error {
	return inv.InvList.SerializeKeep(w, old)
}

func (inv *StorageInv) AllowPut(c *minetest.Client, slot uint32) bool {
	return true
}

func (inv *StorageInv) AllowTake(c *minetest.Client, slot uint32) bool {
	return true
}
