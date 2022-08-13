package utils

import "sync"

// Custom sync.WaitGroup equivalent, for testing purposes
// Has counter visible for easy monitoring of goroutines and Names for what goroutine is still running
// To create WaitGroupCounter, call NewWaitGroupCounter()
type WaitGroupCounter struct {
	WaitGroup *sync.WaitGroup
	Counter   int
	Names     *sync.Map
}

// Adds a goroutine to the WaitGroup
// delta cannot be negative
// name must be unique
func (w *WaitGroupCounter) Add(delta int, name string) error {
	w.Counter += delta

	if _, ok := w.Names.Load(name); ok {
		return nil
	}

	w.Names.Store(name, true)
	w.WaitGroup.Add(delta)
	return nil
}

// Removes a goroutine from the waitgroup
// name must be the same name provided to Add
func (w *WaitGroupCounter) Done(name string) {
	w.Counter--
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
