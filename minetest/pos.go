package minetest

import (
	"github.com/anon55555/mt"

	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/g3n/engine/math32"
	"io"
	"math"
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

type PosUpdater func(c *Client, pos *ClientPos, lu time.Duration)

var (
	posUpdatersMu sync.RWMutex
	posUpdaters   []Registerd[PosUpdater]
)

// PosUpdater is called with a UNLOCKED ClientPos
func RegisterPosUpdater(pu PosUpdater) {
	posUpdatersMu.Lock()
	defer posUpdatersMu.Unlock()

	posUpdaters = append(posUpdaters, Registerd[PosUpdater]{Caller(1), pu})
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

		//calc dtime
		now := time.Now()
		dtime := now.Sub(cpos.LastUpdate)

		func() {
			cpos.Lock()
			defer cpos.Unlock()

			//anitcheat
			newpos := PlayerPos2PPos(pp, cpos.CurPos.Dim)
			if !anticheatPos(c, cpos.CurPos, newpos, dtime) {
				c.Logf("client moved to fast!\n")

				c.SendCmd(&mt.ToCltMovePlayer{
					Pos: cpos.CurPos.Pos.Pos,

					Pitch: cpos.CurPos.Pitch,
					Yaw:   cpos.CurPos.Yaw,
				})
				return
			}

			//updatepos
			cpos.LastUpdate = now
			cpos.OldPos = cpos.CurPos
			cpos.CurPos = newpos
		}()

		for _, u := range posUpdaters {
			timeout := makePktTimeout()
			done := make(chan struct{})

			go func() {
				u.Thing(c, cpos, dtime)
				close(done)
			}()

			select {
			case <-done:
				timeout.Stop()
			case <-timeout.C:
				c.Logf("Timeout waiting for posUpdater! registerd at %s\n", u.Path)
			}
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

// Returns distance in 10th nodes
// speed is < 0, no max speed
func MaxSpeed(clt *Client) float64 {
	return 100 //TODO
}

// Check if NewPos is valid
func anticheatPos(clt *Client, old, new PPos, dtime time.Duration) bool {
	speed := MaxSpeed(clt)
	if speed < 0 {
		return true
	}

	curspeed := Distance(old.Pos.Pos, new.Pos.Pos) / dtime.Seconds()

	return curspeed < speed
}

func Distance(a, b [3]float32) float64 {
	var number float32

	number += math32.Pow((a[0] - b[0]), 2)
	number += math32.Pow((a[1] - b[1]), 2)
	number += math32.Pow((a[2] - b[2]), 2)

	return math.Sqrt(float64(number))
}
