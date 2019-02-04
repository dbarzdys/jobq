package jobq

import (
	"context"
)

// Job handles tasks consumed from database.
// This interface should be implemented and
// registered using jobq.Manager.
type Job interface {
	HandleTask(context.Context, *Task) error
}
