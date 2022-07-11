package utils

import (
	"testing"
	"time"
)

func TestWaitGroupCounterAdd(t *testing.T) {
	wg := NewWaitGroupCounter()

	err := wg.Add(1, "TestWaitGroupCounterAdd")

	if err != nil {
		t.Fail()
	}

	if wg.Counter != 1 {
		t.FailNow()
	}
}

func TestWaitGroupCounterDone(t *testing.T) {
	wg := NewWaitGroupCounter()

	err := wg.Add(1, "TestWaitGroupCounterDone")

	if err != nil {
		t.Fail()
	}

	wg.Done("TestWaitGroupCounterDone")
	if wg.Counter != 0 {
		t.FailNow()
	}
}
func TestWaitGroupCounterWait(t *testing.T) {
	wg := NewWaitGroupCounter()

	err := wg.Add(1, "TestWaitGroupCounterWait")

	if err != nil {
		t.Fail()
	}

	go func() {
		time.Sleep(time.Duration(1000))
		wg.Done("TestWaitGroupCounterWait")
	}()

	wg.Wait()

	if wg.Counter != 0 {
		t.FailNow()
	}
}
