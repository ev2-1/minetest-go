package minetest

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidLocation   = errors.New("the location specified is Invalid")
	ErrInvalidStack      = errors.New("tried to interact with invalid stack")
	ErrInvalidInv        = errors.New("tried to interact with invalid inventory")
	ErrStackInsufficient = errors.New("stack quanitiy is ErrStackInsufficient")
	ErrStackNotEmpty     = errors.New("the stack interacted with is not empty")
	ErrInvalidPos        = errors.New("the position supplied is not parsable (e.g. not 3 dimensional)")
	ErrOutOfSpace        = errors.New("inventory already full")
)

type nodeDefNotFoundErr struct {
	name string
}

func (ndnfe nodeDefNotFoundErr) Error() string {
	return fmt.Sprintf("Node definition not Found: '%s'", ndnfe.name)
}
