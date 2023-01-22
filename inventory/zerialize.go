package inventory

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

type ErrInvalidInvAction struct {
	Action string
}

func (err *ErrInvalidInvAction) Error() string {
	return fmt.Sprintf("Invalid Inventory Action '%s'!", err.Action)
}

// Reads space or colon speperated strings from io.Reader
func ReadString(r io.Reader, colon bool) (str string) {
	var p = make([]byte, 1)

	for {
		_, err := r.Read(p)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			} else {
				panic(SerializationError{err})
			}
		}

		if p[0] == " "[0] || (colon && p[0] == ":"[0]) {
			return
		} else {
			str += string(p[0])
		}
	}
}

// Reads space speperated uint16 from io.Reader
func ReadUint16(r io.Reader, colon bool) (i uint16) {
	str := ReadString(r, colon)
	ii, err := strconv.ParseUint(str, 10, 16)
	if err != nil {
		panic(SerializationError{err})
	}

	return uint16(ii)
}

func ReadInt(r io.Reader, colon bool) (i int) {
	str := ReadString(r, colon)
	ii, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		panic(SerializationError{err})
	}

	return int(ii)
}

// Stolen from https://github.com/anon55555/mt/blob/bcc58cb3048faa146ed0f90b330ebbe791d53b5c/zerialize.go#L26
type Deserializer interface {
	Deserialize(r io.Reader)
}

func Deserialize(r io.Reader, d interface{}) error {
	return pcall(func() { d.(Deserializer).Deserialize(r) })
}

type SerializationError struct {
	error
}

func pcall(f func()) (rerr error) {
	defer func() {
		switch r := recover().(type) {
		case SerializationError:
			rerr = r.error

		case nil:
		default:
			panic(r)
		}
	}()

	f()

	return
}
