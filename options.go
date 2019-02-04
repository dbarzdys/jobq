package jobq

import (
	"time"
)

// JobOptions contains all job options
type JobOptions struct {
	timeoutEnabled bool
	timeout        time.Duration
	retries        uint
	requeuing      bool
	workerPoolSize uint
}

var defaultJobOptions = JobOptions{
	timeoutEnabled: true,
	timeout:        time.Second * 5,
	retries:        5,
	requeuing:      true,
	workerPoolSize: 1,
}

// JobOption configures job
type JobOption func(*JobOptions)

// WithJobTimeout sets timeout duration that will be used
// if job task fails and has no more retries(default: 5s)
func WithJobTimeout(timeout time.Duration) JobOption {
	return func(opts *JobOptions) {
		opts.timeoutEnabled = true
		opts.timeout = timeout
	}
}

// WithJobTimeoutDisabled disables timeout duration
func WithJobTimeoutDisabled() JobOption {
	return func(opts *JobOptions) {
		opts.timeoutEnabled = false
	}
}

// WithJobRequeueRetries number of retries that will be set
// if job task is requeued (default: 5)
func WithJobRequeueRetries(retries uint) JobOption {
	return func(opts *JobOptions) {
		opts.retries = retries
	}
}

// WithJobRequeuing enables or disables requeuing if
// job task fails (default: true)
func WithJobRequeuing(enabled bool) JobOption {
	return func(opts *JobOptions) {
		opts.requeuing = enabled
	}
}

// WithJobWorkerPoolSize sets how many workers should
// handle this job(default: 1)
func WithJobWorkerPoolSize(size uint) JobOption {
	return func(opts *JobOptions) {
		opts.workerPoolSize = size
	}
}

// TaskOptions contains all task options
type TaskOptions struct {
	startAt        time.Time
	startAtEnabled bool
	retries        uint
}

var defaultTaskOptions = TaskOptions{
	startAtEnabled: false,
	retries:        5,
}

// TaskOption configres task
type TaskOption func(*TaskOptions)

// WithTaskStartTime enables and sets time when task should be executed (default: disabled)
func WithTaskStartTime(t time.Time) TaskOption {
	return func(opts *TaskOptions) {
		opts.startAt = t
		opts.startAtEnabled = true
	}
}

// WithTaskStartTimeDisabled disables task start time
func WithTaskStartTimeDisabled() TaskOption {
	return func(opts *TaskOptions) {
		opts.startAtEnabled = false
	}
}

// WithTaskRetries sets initial retry number for failed tasks (default: 5)
func WithTaskRetries(retries uint) TaskOption {
	return func(opts *TaskOptions) {
		opts.retries = retries
	}
}
