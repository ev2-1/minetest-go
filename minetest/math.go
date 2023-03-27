package minetest

import (
	"math"
)

// Todo implement all functions

func Sin32(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

func Cos32(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

func Sqrt32(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

func Round32(x float32) float32 {
	return float32(math.Round(float64(x)))
}

func RoundToEven32(x float32) float32 {
	return float32(math.RoundToEven(float64(x)))
}
