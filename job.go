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

// JobMiddleware is used to wrap Jobs with middlewares
type JobMiddleware func(Job) Job
