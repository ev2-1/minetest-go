package ao

import (
	"github.com/anon55555/mt"
	"github.com/ev2-1/minetest-go/minetest"

	"sync"
	"time"
)

type PlayerMaker func(clt *minetest.Client, id mt.AOID) ActiveObject

var (
	playerAOmaker   = make(map[string]PlayerMaker)
	playerAOmakerMu sync.RWMutex
)

func RegisterPlayerMaker(name string, mk PlayerMaker) {
	playerAOmakerMu.Lock()
	defer playerAOmakerMu.Unlock()

	playerAOmaker[name] = mk
}

func init() {
	minetest.RegisterLeaveHook(func(l *minetest.Leave) {
		go func() {
			cd := GetClientData(l.Client)
			cd.RLock()
			defer cd.RUnlock()

			RmAO(cd.AOID)
		}()
	})

	minetest.RegisterJoinHook(initPlayerAO)

	minetest.RegisterPosUpdater(func(clt *minetest.Client, p *minetest.ClientPos, dt time.Duration) {
		a := GetPAO(clt)
		if a == nil {
			return
		}

		a.SetPos(p.CurPos)
	})
}

// initPlayerAO initializes a players ActiveObject
func initPlayerAO(clt *minetest.Client) {
	id := getAOID()

	//get playerAOmaker:
	mk, _ := GetPlayerAOmaker(clt)

	//initialize self
	ao := mk(clt, 0)
	ack, err := clt.SendCmd(&mt.ToCltAORmAdd{
		Add: []mt.AOAdd{
			mt.AOAdd{
				ID:       0,
				InitData: ao.AOInit(clt).AOInitData(0),
			},
		},
	})
	if err != nil {
		clt.Fatalf("Error encountered: %s\n", err)
		return
	}

	//for the others
	ao = mk(clt, id)

	// set client to ignore AOID
	cd := GetClientData(clt)

	cd.Lock()
	cd.AOID = id
	cd.Unlock()

	registerAO(ao)

	go func() {
		<-ack // wait for package to be acked
		cd := GetClientData(clt)

		cd.Lock()
		cd.Ready = true
		cd.Unlock()
	}()

	clt.Logf("Registered with AOID %d", id)
}

func GetPlayerAOmaker(clt *minetest.Client) (PlayerMaker, string) {
	cd, ok := clt.GetData("aomaker")
	if !ok {
		return playerAOmaker[minetest.GetConfigStringV("default-aomaker", "debug")], minetest.GetConfigStringV("default-aomaker", "debug")
	}

	name, ok := cd.(string)
	if !ok {
		return playerAOmaker[minetest.GetConfigStringV("default-aomaker", "debug")], minetest.GetConfigStringV("default-aomaker", "debug")
	}

	mk, ok := playerAOmaker[name]
	if !ok {
		return playerAOmaker[minetest.GetConfigStringV("default-aomaker", "debug")], minetest.GetConfigStringV("default-aomaker", "debug")
	}

	return mk, name
}
