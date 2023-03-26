package minetest

import (
	"github.com/anon55555/mt"

	"sync"
	"time"
)

var initTicksMu sync.Once
var ticker <-chan time.Time
var tickDuration, _ = time.ParseDuration("0.05s") // can be changed using arguments

func initTicks() {
	initTicksMu.Do(func() {
		go func() {
			ticker = time.Tick(tickDuration)
			for {
				<-ticker

				go func() {
					physHooksMu.Lock()
					now := float32(time.Now().UnixMilli() / 1000)
					dtime := now - physHooksLast

					for h := range physHooks {
						go h.Thing(dtime)
					}

					physHooksLast = now
					physHooksMu.Unlock()
				}()

				tickHooksMu.RLock()
				for h := range tickHooks {
					h.Thing()
				}
				tickHooksMu.RUnlock()

				pktTickHooksMu.RLock()
				for h := range pktTickHooks {
					h.Thing()
				}
				pktTickHooksMu.RUnlock()
			}
		}()
	})
}

// ao
const AOActionTimeout time.Duration = time.Second

func init() {
	RegisterTickHook(func() {
		duration := <-updateAOs()

		if duration > time.Millisecond*64 {
			Loggers.Warnf("AO tick took %s!\n", 1, duration)
		}
	})
}

func updateAOs() <-chan time.Duration {
	startTime := time.Now()
	ch := make(chan time.Duration)

	go func() {
		for clt := range Clts() {
			err := updateAOsClt(clt)
			if err != nil {
				if ConfigVerbose() {
					clt.Logf("[INFO] err while updating AOs! (%s) (can be normal during join)\n", err)
				}
			}
		}

		ch <- time.Now().Sub(startTime)
	}()

	return ch
}

func updateAOsClt(clt *Client) error {
	ActiveObjectsMu.RLock()
	defer ActiveObjectsMu.RUnlock()

	cd := clt.AOData
	if cd == nil {
		return ErrClientDataNil
	}

	addQueue := make(map[mt.AOID]*AOInit)
	rmQueue := make(map[mt.AOID]struct{})

	cd.RLock()
	defer cd.RUnlock()
	if !cd.Ready {
		return ErrClientNotReady
	}

	//Look for globally added:
	for id, ao := range ActiveObjects {
		if _, ok := cd.AOs[id]; !ok && id != cd.AOID && Relevant(ao, clt) {
			clt.Logf("scheduling %d for aoadd\n", id)
			addQueue[id] = ao.AOInit(clt)
		}
	}

	//Look for globally removed:
	for id, _ := range cd.AOs {
		ao, ok := ActiveObjects[id]
		if (!ok || !Relevant(ao, clt)) && id != cd.AOID {
			clt.Logf("scheduling %d for aorm\n", id)
			rmQueue[id] = struct{}{}
		}
	}

	laq := len(addQueue)

	//skip
	if laq == 0 && len(rmQueue) == 0 {
		return nil
	}

	adds := make([]mt.AOAdd, laq)

	if laq > 0 {
		var i int
		for id, init := range addQueue {
			adds[i] = mt.AOAdd{
				ID:       id,
				InitData: init.AOInitData(id),
			}

			i++
		}
	}

	ack, err := clt.SendCmd(&mt.ToCltAORmAdd{
		Remove: Map2Slice(rmQueue),
		Add:    adds,
	})

	if err != nil {
		clt.Logf("[WARN] Error encounterd when sending pkt: %s\n", err)
		return err
	}

	timeout := time.After(AOActionTimeout)

	select {
	case <-timeout:
		clt.Logf("[WARN] AOAction timed out after %s\n", AOActionTimeout)
		return ErrAOTimeout

	case <-ack:
		//TODO: check which has prio in MT code
		// apply to ClientData
		for id := range addQueue {
			cd.AOs[id] = struct{}{}
		}

		for id := range rmQueue {
			delete(cd.AOs, id)
		}
	}

	return nil
}
