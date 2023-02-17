package minetest

import (
	"github.com/anon55555/mt"

	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
)

// returns dim corosponding to name
// returns nil if Dim does not exist
func GetDim(name string) *Dimension {
	dimensionsMu.RLock()
	defer dimensionsMu.RUnlock()

	dim, ok := dimensions[name]
	if !ok {
		MapLogger.Printf("WARN: dimension %s does not exist!\n", name)
		return nil
	}

	return dim
}

type DimID uint16

func (d DimID) String() string {
	dim := d.Lookup()
	if dim == nil {
		MapLogger.Printf("WARN: dimension %d is not defined!\n", d)
		return fmt.Sprintf("%d", d)
	}

	return dim.Name
}

func (d DimID) Lookup() *Dimension {
	dimensionsMu.RLock()
	defer dimensionsMu.RUnlock()

	name, ok := dimensionsR[d]
	if !ok {
		MapLogger.Printf("WARN: dimension %d has no name!\n", d)
		return nil
	}

	return name
}

func PlayerPos2PPos(pos mt.PlayerPos, d DimID) PPos {
	return PPos{
		Pos{
			Pos:   pos.Pos(),
			Vel:   pos.Vel(),
			Pitch: pos.Pitch(),
			Yaw:   pos.Yaw(),

			Dim: d,
		},

		pos.FOV80,
	}
}

type IntPos struct {
	Pos [3]int16
	Dim DimID
}

type Pos struct {
	Pos, Vel   [3]float32
	Pitch, Yaw float32

	Dim DimID
}

// PPos defines a PlayerPosition
type PPos struct {
	Pos

	FOV80 uint8
}

func (p Pos) IntPos() IntPos {
	return IntPos{
		Pos: p.Int(),
		Dim: p.Dim,
	}
}

func (p Pos) Pos100() (i [3]int32) {
	for k := range i {
		i[k] = int32(p.Pos[k] * 100)
	}

	return
}

func (p Pos) Int() (i [3]int16) {
	for k := range i {
		i[k] = int16(p.Pos[k] / 10)
	}

	return
}

func (p Pos) Vel100() (i [3]int32) {
	for k := range i {
		i[k] = int32(p.Vel[k] * 100)
	}

	return
}

func (p Pos) Pitch100() (i int32) {
	i = int32(p.Pitch * 100)

	return
}

func (p Pos) Yaw100() (i int32) {
	i = int32(p.Yaw * 100)

	return
}

func (p PPos) PlayerPos() mt.PlayerPos {
	return mt.PlayerPos{
		Pos100:   p.Pos100(),
		Vel100:   p.Vel100(),
		Pitch100: p.Pitch100(),
		Yaw100:   p.Yaw100(),

		FOV80: p.FOV80,
	}
}

type ClientPos struct {
	sync.RWMutex

	CurPos     PPos
	OldPos     PPos
	LastUpdate time.Time
}

var be = binary.BigEndian

func (cp *ClientPos) Serialize(w io.Writer) (err error) {
	return binary.Write(w, be, cp.CurPos.Pos)
}

func (cp *ClientPos) Deserialize(w io.Reader) (err error) {
	err = binary.Read(w, be, &cp.CurPos.Pos)
	if err != nil {
		return err
	}

	cp.CurPos.Pos.Pos[1] += 10

	return
}

var posUpdatersMu sync.RWMutex
var posUpdaters []func(c *Client, pos *ClientPos, lu time.Duration)

// PosUpdater is called with a UNLOCKED ClientPos
func RegisterPosUpdater(pu func(c *Client, pos *ClientPos, lu time.Duration)) {
	posUpdatersMu.Lock()
	defer posUpdatersMu.Unlock()

	posUpdaters = append(posUpdaters, pu)
}

func init() {
	RegisterPktProcessor(func(c *Client, pkt *mt.Pkt) {
		var pp mt.PlayerPos

		switch cmd := pkt.Cmd.(type) {
		case *mt.ToSrvPlayerPos:
			pp = cmd.Pos

		case *mt.ToSrvInteract:
			pp = cmd.Pos

		default:
			return
		}

		c.PosState.RLock()
		defer c.PosState.RUnlock()

		cpos := GetFullPos(c)
		cpos.Lock()

		now := time.Now()
		dtime := now.Sub(cpos.LastUpdate)

		cpos.LastUpdate = now
		cpos.OldPos = cpos.CurPos
		cpos.CurPos = PlayerPos2PPos(pp, cpos.CurPos.Dim)
		cpos.Unlock()

		for _, u := range posUpdaters {
			u(c, cpos, dtime)
		}
	})
}

func MakePos(c *Client) *ClientPos {
	_, f, l, _ := runtime.Caller(1)
	c.Logf("makepos %s:%d\n", f, l)

	return &ClientPos{
		CurPos:     PPos{Pos: Pos{Pos: [3]float32{105, 155, 140}}},
		LastUpdate: time.Now(),
	}
}

func GetPos(c *Client) PPos {
	pos := GetFullPos(c)
	pos.RLock()
	defer pos.RUnlock()

	return pos.CurPos
}

// GetFullPos returns pos of player / client
func GetFullPos(c *Client) *ClientPos {
	cd, ok := c.GetData("pos")
	if !ok {
		c.Logf("Info: !ok %T\n", cd)
		cd = MakePos(c)
		c.SetData("pos", cd)
	}

	pos, ok := cd.(*ClientPos)
	if !ok {
		c.Logf("Info: !*ClientPos")
		dat, ok := cd.(*ClientDataSaved)
		if !ok {
			c.Logf("Err: !*ClientDataSaved")
			pos = MakePos(c)
			c.SetData("pos", pos)
			return pos
		}

		pos = new(ClientPos)
		err := pos.Deserialize(bytes.NewReader(dat.Bytes()))
		if err != nil {
			c.Logf("Error while Deserializing ClientPos: %s\n", err)
			pos = MakePos(c)
			c.SetData("pos", pos)
			return pos
		}

		c.SetData("pos", pos)
	}

	return pos
}

// SetPos sets position
// returns old position
func SetPos(c *Client, p PPos, send bool) PPos {
	cpos := GetFullPos(c)
	cpos.Lock()
	defer cpos.Unlock()

	cpos.OldPos = cpos.CurPos
	cpos.CurPos = p

	if cpos.OldPos.Dim != cpos.CurPos.Dim {
		c.PosState.Lock()
		defer c.PosState.Unlock()

		c.Logf("Leaving DIM %s (%d); joining DIM %s (%d)\n",
			cpos.OldPos.Dim, cpos.OldPos.Dim,
			cpos.CurPos.Dim, cpos.CurPos.Dim,
		)

		// TODO Black screen

		now := time.Now()

		//unloading all old blks:
		_, err := unloadAll(c)
		if err != nil {
			c.Logf("Failed to switch dimension %s: %s\n", cpos.CurPos.Dim, err)
		}

		c.Logf("Waiting for clt to ack all overwrites\n")
		time.Sleep(time.Millisecond * 64)
		//<-ack // TODO

		c.Logf("switched dimensions in %s\n",
			time.Now().Sub(now),
		)
	}

	// send client own pos if required
	if send {
		c.SendCmd(&mt.ToCltMovePlayer{
			Pos:   p.Pos.Pos,
			Pitch: p.Pitch,
			Yaw:   p.Yaw,
		})
	}

	return cpos.OldPos
}
