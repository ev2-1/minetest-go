package minetest

import (
	"reflect"
)

func compareType(a, b interface{}) bool {
	return reflect.TypeOf(a) == reflect.TypeOf(b)
}
