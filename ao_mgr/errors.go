package ao

import "errors"

var (
	ErrClientDataNil = errors.New("client data unexpectedly nil")
	ErrAOTimeout     = errors.New("ActiveObject timeout reached")
)
