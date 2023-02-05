package inventory

import (
	"github.com/ev2-1/minetest-go/minetest"
)

// Give adds cnt itms to inv as seen out of cs view.
// If InvLocation.Stack is <0 the function will try to figure out a free slot
// Returns the about of items added and any given error
func Give(c *minetest.Client, inv *InvLocation, cnt uint16, itm string) (uint16, <-chan struct{}, error) {
	defer c.Logger.Printf("ret")
	c.Logf("1")

	// aquire inv
	rwinv, err := inv.Aquire(c)
	if err != nil {
		return 0, nil, err
	}
	c.Logf("2")

	rwinv.Lock()
	defer func() {
		str, err := SerializeString(rwinv.Serialize)
		if err != nil {
			panic(err)
		}

		c.Logger.Printf("sendupdate")
		inv.SendUpdate(str, c)
	}()
	defer rwinv.Unlock()

	var added uint16

	// find free / usable slot
	if inv.Stack >= 0 {
		c.Logf("stack>=0")
		// stack is specified
		list, ok := rwinv.Get(inv.Name)
		if !ok {
			return added, nil, ErrInvalidStack
		}

		s, ok := list.GetStack(inv.Stack)
		if !ok {
			return added, nil, ErrInvalidStack
		}

		if !(s.Count == 0 || s.Name == itm) {
			return added, nil, ErrStackNotEmpty
		}

		s.Name = itm
		oldcount := s.Count
		if s.Count+cnt < s.Count { // check if it would overflow
			s.Count = 65535
			added = s.Count - oldcount
		} else {
			s.Count += cnt
		}

		ok = list.SetStack(inv.Stack, s)
		if !ok {
			return 0, nil, ErrInvalidStack
		}

		if added == cnt {
			return added, nil, nil
		} else {
			return added, nil, ErrOutOfSpace
		}
	}
	c.Logf("3")

	// add item to first one
	list, ok := rwinv.Get(inv.Name)
	if !ok {
		return added, nil, ErrInvalidStack
	}
	c.Logf("4")

	for i := 0; i < list.Width(); i++ {
		c.Logger.Printf("wi: %d", i)

		stack, ok := list.GetStack(i)
		if !ok {
			c.Logf("!ok")
			continue
		}

		if !(stack.Count == 0 || stack.Name == itm) {
			c.Logf("!count==0")
			continue
		}

		stack.Name = itm
		oldcount := stack.Count
		if stack.Count+cnt < stack.Count { // check if it would overflow
			c.Logf("overflow")
			stack.Count = 65535

			added += stack.Count - oldcount
			cnt -= stack.Count - oldcount
		} else {
			c.Logf("not overflow")
			stack.Count += cnt
			added += cnt
		}

		c.Logf("fix stack, %d, %+#v", i, stack)
		ok = list.SetStack(i, stack)
		if !ok {
			return 0, nil, ErrInvalidStack
		}

		if added == cnt {
			return added, nil, nil
		}
	}

	return added, nil, nil
}