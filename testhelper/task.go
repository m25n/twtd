package testhelper

import (
	"context"
	"github.com/m25n/twt/task"
)

func SyncEnqueueTask(_ context.Context, task task.Task) error {
	task(context.Background())
	return nil
}

func NoopEnqueueTask(_ context.Context, _ task.Task) error {
	return nil
}

func StubEnqueueTask(err error) task.EnqueueFunc {
	return func(_ context.Context, _ task.Task) error {
		return err
	}
}
