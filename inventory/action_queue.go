package inventory

import (
	"sync"
)

var (
	queueOnce sync.Once
	queue     = make(chan InvAction, 100)
)

func ensureQueue() {
	queueOnce.Do(func() {
		go queueFunc()
	})
}

func queueFunc() {
	var act InvAction

	for {
		act = <-queue

		act.Apply()
	}
}

func AppendQueue(acts ...InvAction) {
	ensureQueue()

	for _, act := range acts {
		queue <- act
	}
}


