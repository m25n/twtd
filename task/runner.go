package task

import (
	"context"
	"errors"
	"time"
)

type Task func(context.Context)

type EnqueueFunc func(ctx context.Context, task Task) error

type Runner struct {
	tasks   chan Task
	workers []*worker
}

func NewRunner(numWorkers int) *Runner {
	tasks := make(chan Task)
	workers := make([]*worker, numWorkers)
	for i := 0; i < numWorkers; i++ {
		workers[i] = newWorker(tasks)
	}
	return &Runner{tasks: tasks, workers: workers}
}

var EnqueuingTimeoutErr = errors.New("timeout enqueuing task")

func (r *Runner) Enqueue(ctx context.Context, task Task) error {
	select {
	case r.tasks <- task:
		return nil
	case <-ctx.Done():
		return EnqueuingTimeoutErr
	}
}

func (r *Runner) Stop() {
	for _, w := range r.workers {
		w.Stop()
	}
}

type worker struct {
	tasks chan Task
	stop  chan struct{}
}

func newWorker(tasks chan Task) *worker {
	w := &worker{tasks: tasks, stop: make(chan struct{})}
	go w.loop()
	return w
}

func (w *worker) loop() {
	for {
		select {
		case task := <-w.tasks:
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			task(ctx)
			cancel()
		case <-w.stop:
			return
		}
	}
}

func (w *worker) Stop() {
	w.stop <- struct{}{}
}
