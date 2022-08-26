package minetest

import (
	"sync"
)

var (
	stage1     []func()
	stage1Mu   sync.RWMutex
	stage1Once sync.Once
)

func RegisterStage1(f func()) {
	stage1Mu.Lock()
	defer stage1Mu.Unlock()

	stage1 = append(stage1, f)
}

func Stage1() {
	stage1Once.Do(func() {
		stage1Mu.RLock()
		defer stage1Mu.RUnlock()

		for _, f := range stage1 {
			f()
		}

		// gets deleted to save space
		stage1 = nil
	})
}

var (
	stage2     []func()
	stage2Mu   sync.RWMutex
	stage2Once sync.Once
)

func RegisterStage2(f func()) {
	stage2Mu.Lock()
	defer stage2Mu.Unlock()

	stage2 = append(stage2, f)
}

func Stage2() {
	stage2Once.Do(func() {
		stage2Mu.RLock()
		defer stage2Mu.RUnlock()

		for _, f := range stage2 {
			f()
		}

		// gets deleted to save space
		stage2 = nil
	})
}
