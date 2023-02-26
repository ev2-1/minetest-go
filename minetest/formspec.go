package minetest

import (
	"github.com/anon55555/mt"

	"sync"
	"time"
)

var (
	formspecsMu sync.RWMutex
	formspecs   = make(map[string]*Registerd[FormspecDef])
)

type FormSpecSubmitFunc func(c *Client, values map[string]string, edittime time.Duration, closed bool)

type FormspecDef struct {
	Name   string
	Spec   string
	Submit FormSpecSubmitFunc
}

func RegisterFormspec(spec FormspecDef) {
	formspecsMu.Lock()
	defer formspecsMu.Unlock()

	formspecs[spec.Name] = &Registerd[FormspecDef]{Caller(1), spec}
}

// Returns formspec if registerd
// returns nil if not
func GetSpec(name string) (spec *Registerd[FormspecDef]) {
	formspecsMu.RLock()
	defer formspecsMu.RUnlock()

	return formspecs[name]
}

// name is name of registerd FormspecDef
// returns ErrInvalidFormspec if formspec is not registered
func (c *Client) ShowSpec(name string) (<-chan struct{}, error) {
	rspec := GetSpec(name)
	if rspec == nil {
		return nil, ErrInvalidFormspec
	}

	ack, err := c.SendCmd(&mt.ToCltShowFormspec{
		Formspec: rspec.Thing.Spec,
		Formname: name,
	})

	if err != nil {
		return nil, err
	}

	go func() {
		<-ack
		c.openSpecsMu.Lock()
		defer c.openSpecsMu.Unlock()

		c.openSpecs[name] = time.Now()
	}()

	return ack, err
}

func init() {
	RegisterPktProcessor(func(c *Client, pkt *mt.Pkt) {
		cmd, ok := pkt.Cmd.(*mt.ToSrvInvFields)
		if !ok {
			return
		}

		def := GetSpec(cmd.Formname)
		if def == nil {
			c.Logf("Client submitted for unknown formspec '%s'\n", cmd.Formname)
			return
		}

		//fieldsMap:
		m := make(map[string]string)
		for _, field := range cmd.Fields {
			m[field.Name] = field.Value
		}

		if def.Thing.Submit == nil {
			c.Logf("Spec registerd at '%s' is nil\n", def.Path())
			return
		}

		c.openSpecsMu.Lock()
		defer c.openSpecsMu.Unlock()
		var t time.Duration = -1

		oldtime, ok := c.openSpecs[cmd.Formname]
		if ok {
			t = time.Now().Sub(oldtime)
		}

		//unopen for clt
		if m["quit"] == "true" {
			delete(c.openSpecs, cmd.Formname)
			def.Thing.Submit(c, m, t, true)
			return
		}

		def.Thing.Submit(c, m, t, false)
	})
}
