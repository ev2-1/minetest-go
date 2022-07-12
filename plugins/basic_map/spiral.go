package main

type pos [3]int16

func (p pos) add(o pos) pos {
	return pos{
		p[0] + o[0],
		p[1] + o[1],
		p[2] + o[2],
	}
}

func spiral(n int16) []pos {
	o := pos{-n / 2, 0, -n / 2}

	var left, top, right, bottom int16 = 0, 0, n - 1, n - 1
	sz := n * n
	s := make([]pos, sz)
	i := sz - 1
	for left < right {
		// work right, along top
		for c := left; c <= right; c++ {
			s[i] = pos{top, 0, c}.add(o)
			i--
		}
		top++
		// work down right side
		for r := top; r <= bottom; r++ {
			s[i] = pos{r, 0, right}.add(o)
			i--
		}
		right--
		if top == bottom {
			break
		}
		// work left, along bottom
		for c := right; c >= left; c-- {
			s[i] = pos{bottom, 0, c}.add(o)
			i--
		}
		bottom--
		// work up left side
		for r := bottom; r >= top; r-- {
			s[i] = pos{r, 0, left}.add(o)
			i--
		}
		left++
	}
	// center (last) element
	s[i] = pos{top, 0, left}.add(o)

	return s
}
