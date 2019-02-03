package jobq

import (
	"context"
)

type Job interface {
	HandleTask(context.Context, *Task) error
}
