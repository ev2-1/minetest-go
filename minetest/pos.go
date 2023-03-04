package minetest

import (
	"github.com/anon55555/mt"

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

func (cp *ClientPos) Copy() ClientPos {
	cp.RLock()
	defer cp.RUnlock()

	if cp == nil {
		return ClientPos{}
	} else {
		return ClientPos{
			CurPos:     cp.CurPos,
			OldPos:     cp.OldPos,
			LastUpdate: cp.LastUpdate,
		}
	}
}

func (cp *ClientPos) Serialize(w io.Writer) (err error) {
	if cp == nil {
		return ErrNilValue
	}

	return binary.Write(w, be, cp.CurPos)
}

func (cp *ClientPos) Deserialize(w io.Reader) (err error) {
	err = binary.Read(w, be, &cp.CurPos)
	if err != nil {
		return err
	}

	return
}

type PosUpdater func(c *Client, pos *ClientPos, lu time.Duration)

var (
	posUpdatersMu sync.RWMutex
	posUpdaters   = make(map[*Registerd[PosUpdater]]struct{})
)

// PosUpdater is called with a UNLOCKED ClientPos
func RegisterPosUpdater(pu PosUpdater) HookRef[Registerd[PosUpdater]] {
	posUpdatersMu.Lock()
	defer posUpdatersMu.Unlock()

	r := &Registerd[PosUpdater]{Caller(1), pu}
	ref := HookRef[Registerd[PosUpdater]]{&posUpdatersMu, posUpdaters, r}

	posUpdaters[r] = struct{}{}

	return ref
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

		cpos := c.GetFullPos()

		//calc dtime
		now := time.Now()
		dtime := now.Sub(cpos.LastUpdate)

		func() {
			cpos.Lock()
			defer cpos.Unlock()

			//anitcheat
			newpos := PlayerPos2PPos(pp, cpos.CurPos.Dim)
			if !AnticheatPos(c, cpos.CurPos, newpos, dtime) {
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

		//copy to prevent deadlock
		posUpdatersMu.RLock()
		updaters := make([]Registerd[PosUpdater], len(posUpdaters))
		var i int
		for k := range posUpdaters {
			updaters[i] = *k

			i++
		}
		posUpdatersMu.RUnlock()

		for u := range posUpdaters {
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
				c.Logf("Timeout waiting for posUpdater! registerd at %s\n", u.Path())
			}
		}

		c.MapLoad()
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

func (c *Client) GetPos() PPos {
	pos := c.GetFullPos()
	pos.RLock()
	defer pos.RUnlock()

	return pos.CurPos
}

// GetFullPos returns pos of player / client
func (c *Client) GetFullPos() *ClientPos {
	return c.Pos
}

// SetPos sets position
// returns old position
func SetPos(c *Client, p PPos, send bool) PPos {
	cpos := c.GetFullPos()
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
func MaxSpeed(clt *Client) float32 {
	return -100 //TODO
}

func (c *Client) PointRange() float32 {
	return 65 + 5 //TODO
}

// PointRange or Default
// Returns pr if pr >= 0 && pr >= default else default
func (c *Client) PRoD(pr float32) float32 {
	d := c.PointRange()

	if pr >= 0 && pr >= d {
		return pr
	} else {
		return d
	}
}

// Check if NewPos is valid
func AnticheatPos(clt *Client, old, new PPos, dtime time.Duration) bool {
	speed := MaxSpeed(clt)
	if speed < 0 {
		return true
	}

	curspeed := Distance(old.Pos.Pos, new.Pos.Pos) / float32(dtime.Seconds())

	return curspeed < speed
}

func Distance(a, b [3]float32) float32 {
	var number float32

	number += math32.Pow((a[0] - b[0]), 2)
	number += math32.Pow((a[1] - b[1]), 2)
	number += math32.Pow((a[2] - b[2]), 2)

	return float32(math.Sqrt(float64(number)))
}
