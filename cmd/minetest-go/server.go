package main

import (
	"github.com/ev2-1/minetest-go/minetest"
)

func main() {
	minetest.Stage1()
	minetest.Stage2()

	// final step
	minetest.Run()
}
