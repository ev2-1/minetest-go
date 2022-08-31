package ao

import (
	"github.com/anon55555/mt"
	"github.com/g3n/engine/math32"
)

func Distance(a, b mt.Vec) float32 {
	var number float32

	number += math32.Pow((a[0] - b[0]), 2)
	number += math32.Pow((a[1] - b[1]), 2)
	number += math32.Pow((a[2] - b[2]), 2)

	return math32.Sqrt(number)
}

func mulVec(v mt.Vec, f float32) mt.Vec {
	return mt.Vec{v[0] * f, v[1] * f, v[2] * f}
}

func addVec(a, b mt.Vec) mt.Vec {
	return mt.Vec{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}

func mulVecs(a, b mt.Vec) mt.Vec {
	return mt.Vec{a[0] * b[0], a[1] * b[1], a[2] * b[2]}
}
