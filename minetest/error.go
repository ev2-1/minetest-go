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

	ErrInvalidFormspec = errors.New("formspec is not registered")

	ErrClientNotReady = errors.New("client not ready")
)

var (
	ErrNilValue = errors.New("unexpected nil value")
)

type nodeDefNotFoundErr struct {
	name string
}

func (ndnfe nodeDefNotFoundErr) Error() string {
	return fmt.Sprintf("Node definition not Found: '%s'", ndnfe.name)
}

type BlamedErr struct {
	Err   error
	Cause string
}

func (err *BlamedErr) Error() string {
	return fmt.Sprintf("%s because of %s!", err.Err, err.Cause)
}

func (r Registerd[T]) Blame(err error) error {
	return &BlamedErr{
		Err:   err,
		Cause: r.Path(),
	}
}
