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
func GetDim(name string) *Dim {
	dimensionsMu.RLock()
	defer dimensionsMu.RUnlock()

	dim, ok := dimensions[name]
	if !ok {
		MapLogger.Printf("WARN: dimension %s does not exist!\n", name)
		return nil
	}

	return &dim
}

type Dim uint16

func (d Dim) String() string {
	dimensionsMu.RLock()
	defer dimensionsMu.RUnlock()

	name, ok := dimensionsR[d]
	if !ok {
		MapLogger.Printf("WARN: dimension %d has no name!\n", d)
		return fmt.Sprintf("%d", d)
	}

	return name
}

func PlayerPos2Pos(pos mt.PlayerPos, d Dim) Pos {
	return Pos{
		Pos:   pos.Pos(),
		Vel:   pos.Vel(),
		Pitch: pos.Pitch(),
		Yaw:   pos.Yaw(),

		FOV80: pos.FOV80,

		Dim: d,
	}
}

type IntPos struct {
	Pos [3]int16
	Dim Dim
}

type Pos struct {
	Pos, Vel   [3]float32
	Pitch, Yaw float32

	FOV80 uint8

	Dim Dim
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

func (p Pos) PlayerPos() mt.PlayerPos {
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

	Pos        Pos
	OldPos     Pos
	LastUpdate time.Time
}

var be = binary.BigEndian

func (cp *ClientPos) Serialize(w io.Writer) (err error) {
	return binary.Write(w, be, cp.Pos)
}

func (cp *ClientPos) Deserialize(w io.Reader) (err error) {
	err = binary.Read(w, be, &cp.Pos)
	if err != nil {
		return err
	}

	cp.Pos.Pos[1] += 10

	return
}

var posUpdatersMu sync.RWMutex
var posUpdaters []func(c *Client, pos *ClientPos, lu time.Duration)

// PosUpdater is called with a LOCKED ClientPos
func RegisterPosUpdater(pu func(c *Client, pos *ClientPos, lu time.Duration)) {
	posUpdatersMu.Lock()
	defer posUpdatersMu.Unlock()

	posUpdaters = append(posUpdaters, pu)
}

func init() {
	RegisterPktProcessor(func(c *Client, pkt *mt.Pkt) {
		pp, ok := pkt.Cmd.(*mt.ToSrvPlayerPos)

		if c.PosState != PsOk {
			c.Logf("PosState != PsOk, waiting")
			<-c.PosUnlock
			c.Logf("PosUnlock")
		}

		if ok {
			cpos := GetFullPos(c)
			cpos.Lock()
			defer cpos.Unlock()

			now := time.Now()
			dtime := now.Sub(cpos.LastUpdate)

			cpos.LastUpdate = now
			cpos.OldPos = cpos.Pos
			cpos.Pos = PlayerPos2Pos(pp.Pos, cpos.Pos.Dim)

			for _, u := range posUpdaters {
				u(c, cpos, dtime)
			}
		}
	})
}

func MakePos(c *Client) *ClientPos {
	_, f, l, _ := runtime.Caller(1)
	c.Logf("makepos %s:%d\n", f, l)

	return &ClientPos{
		Pos:        Pos{Pos: [3]float32{105, 155, 140}},
		LastUpdate: time.Now(),
	}
}

func GetPos(c *Client) Pos {
	pos := GetFullPos(c)

	return pos.Pos
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
func SetPos(c *Client, p Pos) Pos {
	cpos := GetFullPos(c)
	cpos.Lock()
	defer cpos.Unlock()

	cpos.OldPos = cpos.Pos
	cpos.Pos = p

	if cpos.OldPos.Dim != cpos.Pos.Dim {
		unlockchan := make(chan struct{})

		c.PosUnlock = unlockchan
		defer close(unlockchan)
		c.PosState = PsTransition
		defer func() { c.PosState = PsOk }()

		c.Logf("Leaving DIM %s (%d); joining DIM %s (%d)\n",
			cpos.OldPos.Dim, cpos.OldPos.Dim,
			cpos.Pos.Dim, cpos.Pos.Dim,
		)

		now := time.Now()

		//unloading all old blks:
		_, err := unloadAll(c)
		if err != nil {
			c.Logf("Failed to switch dimension %s: %s\n", cpos.Pos.Dim, err)
		}

		c.Logf("Waiting for clt to ack all overwrites\n")
		time.Sleep(time.Millisecond * 64)
		//<-ack // TODO

		c.Logf("switched dimensions in %s\n",
			time.Now().Sub(now),
		)
	}

	return cpos.OldPos
}
