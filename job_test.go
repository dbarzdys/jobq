package jobq

import "context"

type mockJob struct {
	onHandleTask func(context.Context, *Task) error
}

func (j mockJob) HandleTask(ctx context.Context, t *Task) error {
	return j.onHandleTask(ctx, t)
}
