package services

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func waitWithTimeout(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("timed out after %s", timeout)
	}
}

func TestTaskRunnerRespectsConcurrencyLimit(t *testing.T) {
	runner := NewTaskRunner(nil, 2)

	var current int32
	var maxSeen int32
	var wg sync.WaitGroup

	for i := 0; i < 6; i++ {
		wg.Add(1)
		runner.Submit("concurrency_limit", func() {
			defer wg.Done()

			active := atomic.AddInt32(&current, 1)
			for {
				previous := atomic.LoadInt32(&maxSeen)
				if active <= previous || atomic.CompareAndSwapInt32(&maxSeen, previous, active) {
					break
				}
			}

			time.Sleep(25 * time.Millisecond)
			atomic.AddInt32(&current, -1)
		})
	}

	waitWithTimeout(t, &wg, time.Second)

	if got := atomic.LoadInt32(&maxSeen); got > 2 {
		t.Fatalf("expected max concurrency <= 2, got %d", got)
	}
}

func TestTaskRunnerReleasesSemaphoreAfterPanic(t *testing.T) {
	runner := NewTaskRunner(nil, 1)

	firstDone := make(chan struct{})
	secondDone := make(chan struct{})

	runner.Submit("panic_task", func() {
		defer close(firstDone)
		panic("boom")
	})

	select {
	case <-firstDone:
	case <-time.After(time.Second):
		t.Fatal("panic task did not finish in time")
	}

	runner.Submit("follow_up", func() {
		close(secondDone)
	})

	select {
	case <-secondDone:
	case <-time.After(time.Second):
		t.Fatal("follow-up task did not start after panic")
	}
}

func TestTaskRunnerIgnoresNilTask(t *testing.T) {
	runner := NewTaskRunner(nil, 1)
	runner.Submit("nil_task", nil)
}
