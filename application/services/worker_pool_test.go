package services

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPoolRespectsConcurrencyLimit(t *testing.T) {
	pool := NewWorkerPool(nil, 2)
	defer func() {
		_ = pool.Stop(context.Background())
	}()

	var current int32
	var maxSeen int32
	var wg sync.WaitGroup

	for i := 0; i < 6; i++ {
		wg.Add(1)
		if err := pool.Submit("concurrency_limit", func() {
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
		}); err != nil {
			t.Fatalf("submit error: %v", err)
		}
	}

	waitWithTimeout(t, &wg, time.Second)

	if got := atomic.LoadInt32(&maxSeen); got > 2 {
		t.Fatalf("expected max concurrency <= 2, got %d", got)
	}
}

func TestWorkerPoolRejectsNewJobsAfterStop(t *testing.T) {
	pool := NewWorkerPool(nil, 1)
	if err := pool.Stop(context.Background()); err != nil {
		t.Fatalf("stop error: %v", err)
	}

	err := pool.Submit("after_stop", func() {})
	if err == nil {
		t.Fatal("expected submit to fail after stop")
	}
}
