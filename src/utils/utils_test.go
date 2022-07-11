package utils

import (
	"testing"
	"time"
)

func TestWaitGroupCounterAdd(t *testing.T) {
	wg := WaitGroupCounter{}

	wg.Add(1)

	if wg.Counter != 1 {
		t.FailNow()
	}
}

func TestWaitGroupCounterSubtract(t *testing.T) {
	wg := WaitGroupCounter{}

	wg.Add(1)
	wg.Add(-1)
	if wg.Counter != 0 {
		t.FailNow()
	}
}

func TestWaitGroupCounterDone(t *testing.T) {
	wg := WaitGroupCounter{}

	wg.Add(1)
	wg.Done()
	if wg.Counter != 0 {
		t.FailNow()
	}
}
func TestWaitGroupCounterWait(t *testing.T) {
	wg := WaitGroupCounter{}

	wg.Add(1)

	go func() {
		time.Sleep(time.Duration(1000))
		wg.Done()
	}()

	wg.Wait()

	if wg.Counter != 0 {
		t.FailNow()
	}
}