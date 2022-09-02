package basic_init

// This is only a placeholder

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"
)

func init() {
	minetest.RegisterInitHook(func(c *minetest.Client) {
		c.SendCmd(&mt.ToCltPrivs{Privs: []string{"fly", "interact", "fast", "noclip"}})
		c.SendCmd(&mt.ToCltDetachedInv{})
		c.SendCmd(&mt.ToCltMovement{
			DefaultAccel: 2.4,
			AirAccel:     1.2,
			FastAccel:    10,
			WalkSpeed:    4.317,
			CrouchSpeed:  1.295,
			FastSpeed:    30,
			ClimbSpeed:   2.35,
			JumpSpeed:    6.6,
			Fluidity:     1.13,
			Smoothing:    0.5,
			Sink:         23,
			Gravity:      10.4,
		})
		c.SendCmd(&mt.ToCltInvFormspec{
			Formspec: "List main 1\nWidth 5\nEmpty\nEmpty\nEmpty\nEmpty\nEmpty\nEndInventory\n",
		})
		c.SendCmd(&mt.ToCltHP{
			HP: 20,
		})
		c.SendCmd(&mt.ToCltBreath{
			Breath: 10,
		})
	})
}
