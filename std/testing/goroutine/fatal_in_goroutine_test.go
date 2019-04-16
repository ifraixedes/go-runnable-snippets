package goroutine

import (
	"sync"
	"testing"
	"time"
)

/*
This file contains several test functions to see how the testing package behaves
when a test is aborted from a Go routine.
*/

func TestBehaviorFatalInGoroutine(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(time.Second)

		t.Fatal("WANTED ABORTION")
	}()

	wg.Wait()
}

func TestBehaviorSubTestParentFatalInGoroutine(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	t.Run("subtest", func(_ *testing.T) {
		defer wg.Done()
		// NOTE this is not OK
		t.Fatal("WANTED ABORTION")
		time.Sleep(time.Second)
	})

	wg.Wait()
}

func TestBehaviorSubTestFatalInGoroutine(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	t.Run("subtest", func(st *testing.T) {
		defer wg.Done()
		st.Fatal("WANTED ABORTION")
		time.Sleep(time.Second)
	})

	wg.Wait()
}
