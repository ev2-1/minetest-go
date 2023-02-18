package mapLoader

import (
	"github.com/ev2-1/minetest-go/minetest"
)

func loadAround(ppp minetest.IntPos, c *minetest.Client) {
	s := spiral(3)
	pp := pos(ppp)

	for _, p := range s {
		minetest.LoadBlk(c, pp.add(p).IntPos())
	}
	for _, p := range s {
		minetest.LoadBlk(c, pp.add(p).add([3]int16{0, -1, 0}).IntPos())
	}
	for _, p := range s {
		minetest.LoadBlk(c, pp.add(p).add([3]int16{0, 1, 0}).IntPos())
	}

}

type pos minetest.IntPos

func (p pos) IntPos() minetest.IntPos {
	return minetest.IntPos(p)
}

func (p pos) add(o [3]int16) (i pos) {
	i.Dim = p.Dim

	for k := range o {
		i.Pos[k] = p.Pos[k] + o[k]
	}

	return
}

func add(a, b [3]int16) (i [3]int16) {
	for k := range i {
		i[k] = a[k] + b[k]
	}

	return
}

func spiral(n int16) [][3]int16 {
	o := [3]int16{-n / 2, 0, -n / 2}

	var left, top, right, bottom int16 = 0, 0, n - 1, n - 1
	sz := n * n
	s := make([][3]int16, sz)
	i := sz - 1
	for left < right {
		// work right, along top
		for c := left; c <= right; c++ {
			s[i] = add([3]int16{top, 0, c}, o)
			i--
		}
		top++
		// work down right side
		for r := top; r <= bottom; r++ {
			s[i] = add([3]int16{r, 0, right}, o)
			i--
		}
		right--
		if top == bottom {
			break
		}
		// work left, along bottom
		for c := right; c >= left; c-- {
			s[i] = add([3]int16{bottom, 0, c}, o)
			i--
		}
		bottom--
		// work up left side
		for r := bottom; r >= top; r-- {
			s[i] = add([3]int16{r, 0, left}, o)
			i--
		}
		left++
	}
	// center (last) element
	s[i] = add([3]int16{top, 0, left}, o)

	return s
}
