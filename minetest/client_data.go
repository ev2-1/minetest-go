// This file contains abstractions for serializing data
// For how data is stored, and Database access, see client_storage.go
package minetest

import (
	"errors"
	"io"
)

var ErrClientDataNotFound = errors.New("ClientData not found!")
var ErrClientDataInvalidType = errors.New("ClientData has invalid type!")

type ClientData interface {
}

// If a ClientDataSerialize is used as ClientData it will be serialized and saved in the players table
type ClientDataSerialize interface {
	ClientData

	// Serialize should serialize ClientData into some kind of format
	// Binary is prefered as its more space efficient
	Serialize(w io.Writer) (err error)
}

// The raw data loaded from the Database
// implements ClientData
type ClientDataSaved struct {
	data []byte
}

func (c *Client) ensureClientData() {
	if c.data == nil {
		c.data = make(map[string]ClientData)
	}
}

func (c *Client) GetData(field string) (cd ClientData, ok bool) {
	c.dataMu.RLock()
	defer c.dataMu.RUnlock()

	cd, ok = c.data[field]
	return
}

func (c *Client) SetData(field string, value ClientData) (overwrote bool) {
	c.dataMu.Lock()
	defer c.dataMu.Unlock()
	c.ensureClientData()

	_, overwrote = c.data[field]
	c.data[field] = value

	return
}

func (c *Client) DelData(field string) (found bool) {
	c.dataMu.Lock()
	defer c.dataMu.Unlock()

	_, found = c.data[field]
	c.data[field] = nil

	return
}

func (cd *ClientDataSaved) Bytes() (b []byte) {
	b = make([]byte, len(cd.data))

	copy(b, cd.data)

	return
}

func (cd *ClientDataSaved) Serialize(w io.Writer) (err error) {
	_, err = w.Write(cd.data)

	return
}

type ClientDataString struct {
	String string
}

func (cd *ClientDataString) Desc() string {
	return "just a string(TM)"
}

func (cd *ClientDataString) Serialize(w io.Writer) (err error) {
	_, err = w.Write([]byte(cd.String))

	return
}

func TryClientDataString(c *Client, f string) *ClientDataString {
	c.dataMu.Lock()
	defer c.dataMu.Unlock()

	cd := c.data[f]
	if cd == nil {
		return nil
	}

	cds, ok := cd.(*ClientDataString)
	if ok {
		return cds
	}

	d, ok := cd.(*ClientDataSaved)
	if ok {
		cds := ClientDataString{string(d.Bytes())}
		c.data[f] = &cds

		return &cds
	}

	return nil
}

func (cd *ClientDataSaved) Desc() string {
	return "This contains some (not yet) deserialized data!"
}
