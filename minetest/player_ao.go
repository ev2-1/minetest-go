package minetest

import (
	"github.com/anon55555/mt"

	"sync"
	"time"
)

type PlayerAOMaker func(clt *Client, id mt.AOID) ActiveObject

var (
	playerAOmaker   = make(map[string]PlayerAOMaker)
	playerAOmakerMu sync.RWMutex
)

func RegisterPlayerMaker(name string, mk PlayerAOMaker) {
	playerAOmakerMu.Lock()
	defer playerAOmakerMu.Unlock()

	playerAOmaker[name] = mk
}

func init() {
	RegisterLeaveHook(func(l *Leave) {
		go func() {
			cd := l.Client.AOData
			if cd == nil {
				return
			}

			cd.RLock()
			defer cd.RUnlock()

			RmAO(cd.AOID)
		}()
	})

	RegisterJoinHook(initPlayerAO)

	RegisterPosUpdater(func(clt *Client, p *ClientPos, dt time.Duration) {
		a := GetPAO(clt)
		if a == nil {
			return
		}

		a.SetPos(p.CurPos)
	})
}

// initPlayerAO initializes a players ActiveObject
func initPlayerAO(clt *Client) {
	clt.Logf("initializing Player AO\n")

	id := getAOID()

	//get playerAOmaker:
	mk, _ := GetPlayerAOmaker(clt)

	//initialize self
	ao := mk(clt, 0)

	selfinit := ao.AOInit(clt)

	ack, err := clt.SendCmd(&mt.ToCltAORmAdd{
		Add: []mt.AOAdd{
			mt.AOAdd{
				ID:       0,
				InitData: selfinit.AOInitData(0),
			},
		},
	})

	if err != nil {
		clt.Fatalf("Error encountered: %s\n", err)
		return
	}

	//for the others
	ao = mk(clt, id)

	var selfprops mt.AOProps
	for _, msg := range selfinit.AOMsgs {
		if m, ok := msg.(*mt.AOCmdProps); ok {
			selfprops = m.Props

			break
		}
	}

	// set client to ignore AOID
	cd := clt.AOData
	cd.Lock()
	cd.AOID = id
	cd.SelfProps = selfprops
	cd.Unlock()

	registerAO(ao)

	go func() {
		<-ack // wait for package to be acked
		cd := clt.AOData
		cd.Lock()
		defer cd.Unlock()
		cd.Ready = true
	}()

	clt.Logf("Registered with AOID %d", id)
}

func GetPlayerAOmaker(clt *Client) (PlayerAOMaker, string) {
	cd, ok := clt.GetData("aomaker")
	if !ok {
		return playerAOmaker[GetConfigV("default-aomaker", "debug")], GetConfigV("default-aomaker", "debug")
	}

	name, ok := cd.(string)
	if !ok {
		return playerAOmaker[GetConfigV("default-aomaker", "debug")], GetConfigV("default-aomaker", "debug")
	}

	mk, ok := playerAOmaker[name]
	if !ok {
		return playerAOmaker[GetConfigV("default-aomaker", "debug")], GetConfigV("default-aomaker", "debug")
	}

	return mk, name
}
