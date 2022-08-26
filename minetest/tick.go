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

				for _, h := range tickHooks {
					h()
				}

				for _, h := range pktTickHooks {
					h()
				}
			}
		}()
	})
}