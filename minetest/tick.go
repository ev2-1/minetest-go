package minetest

import (
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

					for _, h := range physHooks {
						go h(dtime)
					}

					physHooksLast = now
					physHooksMu.Unlock()
				}()

				tickHooksMu.RLock()
				for _, h := range tickHooks {
					h()
				}
				tickHooksMu.RUnlock()

				pktTickHooksMu.RLock()
				for _, h := range pktTickHooks {
					h()
				}
				pktTickHooksMu.RUnlock()
			}
		}()
	})
}
