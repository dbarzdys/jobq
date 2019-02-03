package jobq

import (
	"time"
)

type JobOptions struct {
	timeoutEnabled bool
	timeout        time.Duration
	retries        uint
	requeuing      bool
	workerPoolSize uint
}

var DefaultJobOptions = JobOptions{
	timeoutEnabled: true,
	timeout:        time.Second * 5,
	retries:        5,
	requeuing:      true,
	workerPoolSize: 1,
}

type JobOption func(*JobOptions)

func WithJobTimeout(timeout time.Duration) JobOption {
	return func(opts *JobOptions) {
		opts.timeoutEnabled = true
		opts.timeout = timeout
	}
}
func WithJobTimeoutDisabled() JobOption {
	return func(opts *JobOptions) {
		opts.timeoutEnabled = false
	}
}

func WithJobRequeueRetries(retries uint) JobOption {
	return func(opts *JobOptions) {
		opts.retries = retries
	}
}

func WithJobRequeuing(requeuing bool) JobOption {
	return func(opts *JobOptions) {
		opts.requeuing = requeuing
	}
}

func WithJobWorkerPoolSize(size uint) JobOption {
	return func(opts *JobOptions) {
		opts.workerPoolSize = size
	}
}

type TaskOptions struct {
	startAt        time.Time
	startAtEnabled bool
	retries        int
}

var DefaultTaskOptions = TaskOptions{
	startAtEnabled: false,
	retries:        5,
}

type TaskOption func(*TaskOptions)

func WithTaskStartTime(t time.Time) TaskOption {
	return func(opts *TaskOptions) {
		opts.startAt = t
		opts.startAtEnabled = true
	}
}
func WithTaskStartTimeDisabled() TaskOption {
	return func(opts *TaskOptions) {
		opts.startAtEnabled = false
	}
}
func WithTaskRetries(retries int) TaskOption {
	return func(opts *TaskOptions) {
		opts.retries = retries
	}
}
