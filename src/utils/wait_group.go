package utils

import "sync"
import "sync/atomic"

// Custom sync.WaitGroup equivalent, for testing purposes
// Has counter visible for easy monitoring of goroutines and Names for what goroutine is still running
// To create WaitGroupCounter, call NewWaitGroupCounter()
type WaitGroupCounter struct {
	WaitGroup *sync.WaitGroup
	Counter   int64
	Names     *sync.Map
}

// Adds a goroutine to the WaitGroup
// delta cannot be negative
// name must be unique
func (w *WaitGroupCounter) Add(delta int64, name string) error {
	if w.Names == nil {
		panic("wg: Names field uninitialized")
	}
	if _, ok := w.Names.Load(name); ok {
		return nil
	}
	atomic.StoreInt64(&w.Counter, atomic.AddInt64(&w.Counter, delta))

	w.Names.Store(name, true)
	w.WaitGroup.Add(int(delta))
	return nil
}

// Removes a goroutine from the waitgroup
// name must be the same name provided to Add
func (w *WaitGroupCounter) Done(name string) {
	atomic.StoreInt64(&w.Counter, atomic.AddInt64(&w.Counter, -1))

	w.Names.Delete(name)

	w.WaitGroup.Done()
}

func (w *WaitGroupCounter) Wait() {
	w.WaitGroup.Wait()
}

func NewWaitGroupCounter() *WaitGroupCounter {
	return &WaitGroupCounter{
		WaitGroup: &sync.WaitGroup{},
		Names:     &sync.Map{},
	}
}
