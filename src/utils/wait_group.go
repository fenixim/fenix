package utils

import "sync"

type WaitGroupCounter struct {
	WaitGroup sync.WaitGroup
	Counter   int
}

func (w *WaitGroupCounter) Add(delta int) {
	w.Counter += delta
	w.WaitGroup.Add(delta)
}

func (w *WaitGroupCounter) Done() {
	w.Counter--
	w.WaitGroup.Done()
}

func (w *WaitGroupCounter) Wait() {
	w.WaitGroup.Wait()
}
