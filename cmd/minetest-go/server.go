package main

import (
	"github.com/ev2-1/minetest-go/minetest"
)

func main() {
	stage1() // generated to init.go by makeinit.sh
	stage2() // generated to init.go by makeinit.sh

	// final step
	minetest.Run()
}
