package help

import (
	"errors"
	"strings"
	"sync"
)

var (
	helps   = make(map[string]*Help)
	helpsMu sync.RWMutex
)

// RegisterHelp registers a Help struct to global map
// if help.Name is empty it is set to name
func RegisterHelp(name string, help *Help) {
	if help.Name == "" {
		help.Name = name
	}

	helpsMu.Lock()
	defer helpsMu.Unlock()

	helps[name] = help
}

// returns a copy of the map containing all Help structs
// does NOT copy Help structs themselves
// entries in the map can be nil
func GetHelps() map[string]*Help {
	helpsMu.RLock()
	defer helpsMu.RUnlock()

	m := make(map[string]*Help, len(helps))

	for k, v := range helps {
		m[k] = v
	}

	return m
}

func GetHelp(name string) *Help {
	helpsMu.RLock()
	defer helpsMu.RUnlock()

	return helps[name]
}

// Help defines the help given for a give chatcommand
type Help struct {
	Name string
	Desc string

	Args []Argument
}

// Argument defines a argument to a chatcommand
type Argument struct {
	Name string
	Type ArgType
}

type ArgType uint8

//go:generate stringer -type ArgType -linecomment
const (
	ArgAny    ArgType = iota //any
	ArgInt                   //int
	ArgUInt                  //uint
	ArgBool                  //bool
	ArgPos                   //pos
	ArgString                //string
)

var (
	ErrInvalidArgType = errors.New("invalid argument type")
)

// ParseArg tries to parse a argument from string
// Syntax is <type>:<name> split at first colon; if no type is specified ArgAny is used
func ParseArg(arg string) (Argument, error) {
	args := strings.SplitN(arg, ":", 2)
	if len(args) == 1 {
		return Argument{
			Name: args[0],
			Type: ArgAny,
		}, nil
	}

	a, ok := argsM[args[0]]
	if !ok {
		return Argument{Name: args[1]}, ErrInvalidArgType
	}

	return Argument{
		Name: args[1],
		Type: a,
	}, nil
}

// ParseArgs tries to parse a space seperated list of arguments
// Syntax is same used by ParseArg
func ParseArgs(args string) ([]Argument, error) {
	if args == "" {
		return nil, nil
	}

	argsS := strings.Split(args, " ")
	a := make([]Argument, len(argsS))
	var err error

	for i, arg := range argsS {
		a[i], err = ParseArg(arg)
		if err != nil {
			return nil, err
		}
	}

	return a, nil
}

// Wraps ParseArgs; panics if error is encountered
func MustParseArgs(args string) []Argument {
	a, err := ParseArgs(args)
	if err != nil {
		panic(err)
	}

	return a
}

var argsM = map[string]ArgType{
	"any":    ArgAny,
	"int":    ArgInt,
	"uint":   ArgUInt,
	"bool":   ArgBool,
	"string": ArgString,
	"pos":    ArgPos,
}
