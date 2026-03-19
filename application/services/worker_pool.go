package services

import (
	"context"
	"errors"
	"sync"

	"github.com/drama-generator/backend/pkg/logger"
)

var errWorkerPoolStopped = errors.New("worker pool stopped")

type workerJob struct {
	name string
	fn   func()
}

type WorkerPool struct {
	log     *logger.Logger
	jobs    chan workerJob
	stopCh  chan struct{}
	wg      sync.WaitGroup
	mu      sync.RWMutex
	stopped bool
}

func NewWorkerPool(log *logger.Logger, concurrency int) *WorkerPool {
	if concurrency <= 0 {
		concurrency = defaultAsyncTaskConcurrency
	}

	pool := &WorkerPool{
		log:    log,
		jobs:   make(chan workerJob, concurrency*4),
		stopCh: make(chan struct{}),
	}

	for i := 0; i < concurrency; i++ {
		pool.wg.Add(1)
		go pool.runWorker()
	}

	return pool
}

func (p *WorkerPool) Submit(name string, fn func()) error {
	if fn == nil {
		return nil
	}

	p.mu.RLock()
	stopped := p.stopped
	p.mu.RUnlock()
	if stopped {
		return errWorkerPoolStopped
	}

	select {
	case <-p.stopCh:
		return errWorkerPoolStopped
	case p.jobs <- workerJob{name: name, fn: fn}:
		return nil
	}
}

func (p *WorkerPool) Stop(ctx context.Context) error {
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		return nil
	}
	p.stopped = true
	close(p.stopCh)
	close(p.jobs)
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		defer close(done)
		p.wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (p *WorkerPool) runWorker() {
	defer p.wg.Done()

	for job := range p.jobs {
		func() {
			defer func() {
				if recovered := recover(); recovered != nil && p.log != nil {
					p.log.Errorw("Worker job panicked", "task", job.name, "panic", recovered)
				}
			}()
			job.fn()
		}()
	}
}
