package minetest

import (
	"github.com/anon55555/mt"
)

func Broadcast(cmd mt.Cmd) []<-chan struct{} {
	acks := []<-chan struct{}{}

	for c := range Clts() {
		a, _ := c.SendCmd(cmd)
		acks = append(acks, a)
	}

	return acks
}
