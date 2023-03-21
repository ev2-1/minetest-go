package minetest

import (
	"github.com/anon55555/mt"

	"fmt"
	"time"
)

type FormSpecSubmitFunc func(c *Client, values map[string]string, edittime time.Duration, closed bool)

type Formspec struct {
	Name   string
	Spec   string
	Submit FormSpecSubmitFunc
}

func (c *Client) RegisterFormspec(spec *Formspec) {
	c.openSpecsMu.Lock()
	defer c.openSpecsMu.Unlock()

	c.openSpecs[spec.Name] = spec
}

// Returns formspec if registerd
// returns nil if not
func (c *Client) GetSpec(name string) (spec *Formspec) {
	c.openSpecsMu.RLock()
	defer c.openSpecsMu.RUnlock()

	return c.openSpecs[name]
}

// name is name of registerd FormspecDef
// returns ErrInvalidFormspec if formspec is not registered
func (c *Client) ShowSpec(spec *Formspec) (<-chan struct{}, error) {
	if spec == nil {
		return nil, ErrInvalidFormspec
	}

	ack, err := c.SendCmd(&mt.ToCltShowFormspec{
		Formspec: spec.Spec,
		Formname: spec.Name,
	})

	if err != nil {
		return nil, err
	}

	go func() {
		<-ack
		c.openSpecsMu.Lock()
		defer c.openSpecsMu.Unlock()

		c.openSpecsT[spec.Name] = time.Now()
		c.openSpecs[spec.Name] = spec
	}()

	return ack, err
}

func (c *Client) ShowSpecf(rspec *Formspec, v ...any) (<-chan struct{}, error) {
	spec := &Formspec{
		Name:   rspec.Name,
		Spec:   fmt.Sprintf(rspec.Spec, v...),
		Submit: rspec.Submit,
	}

	return c.ShowSpec(spec)
}

func init() {
	RegisterPktProcessor(func(c *Client, pkt *mt.Pkt) {
		cmd, ok := pkt.Cmd.(*mt.ToSrvInvFields)
		if !ok {
			return
		}

		def := c.GetSpec(cmd.Formname)
		if def == nil {
			c.Logf("Client submitted for unknown formspec '%s'\n", cmd.Formname)
			return
		}

		//fieldsMap:
		m := make(map[string]string)
		for _, field := range cmd.Fields {
			m[field.Name] = field.Value
		}

		if def.Submit == nil {
			return
		}

		c.openSpecsMu.Lock()
		defer c.openSpecsMu.Unlock()
		var t time.Duration = -1

		oldtime, ok := c.openSpecsT[cmd.Formname]
		if ok {
			t = time.Now().Sub(oldtime)
		}

		//unopen for clt
		if m["quit"] == "true" {
			delete(c.openSpecs, cmd.Formname)
			def.Submit(c, m, t, true)
			return
		}

		def.Submit(c, m, t, false)
	})
}
