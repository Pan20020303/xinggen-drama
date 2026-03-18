package services

import "github.com/drama-generator/backend/pkg/logger"

const defaultAsyncTaskConcurrency = 8

type TaskRunner struct {
	log *logger.Logger
	sem chan struct{}
}

func NewTaskRunner(log *logger.Logger, concurrency int) *TaskRunner {
	if concurrency <= 0 {
		concurrency = defaultAsyncTaskConcurrency
	}
	return &TaskRunner{
		log: log,
		sem: make(chan struct{}, concurrency),
	}
}

func (r *TaskRunner) Submit(name string, fn func()) {
	if fn == nil {
		return
	}

	go func() {
		r.sem <- struct{}{}
		defer func() {
			<-r.sem
			if recovered := recover(); recovered != nil && r.log != nil {
				r.log.Errorw("Async task panicked", "task", name, "panic", recovered)
			}
		}()

		fn()
	}()
}
