package jobq

import (
	"time"
)

// JobOptions contains all job options
type JobOptions struct {
	timeoutEnabled bool
	timeout        time.Duration
	retries        int
	requeuing      bool
	workerPoolSize int
}

func (opts JobOptions) with(args ...JobOption) (JobOptions, error) {
	for _, opt := range args {
		if err := opt(&opts); err != nil {
			return opts, err
		}
	}
	return opts, nil
}

var defaultJobOptions = JobOptions{
	timeoutEnabled: true,
	timeout:        time.Second * 5,
	retries:        5,
	requeuing:      true,
	workerPoolSize: 1,
}

// JobOption configures job
type JobOption func(*JobOptions) error

// WithJobTimeout sets timeout duration that will be used
// if job task fails and has no more retries(default: 5s)
func WithJobTimeout(timeout time.Duration) JobOption {
	return func(opts *JobOptions) error {
		if err := validateTimeout(timeout); err != nil {
			return err
		}
		opts.timeoutEnabled = true
		opts.timeout = timeout
		return nil
	}
}

// WithJobTimeoutDisabled disables timeout duration
func WithJobTimeoutDisabled() JobOption {
	return func(opts *JobOptions) error {
		opts.timeoutEnabled = false
		return nil
	}
}

// WithJobRequeueRetries number of retries that will be set
// if job task is requeued (default: 5)
func WithJobRequeueRetries(retries int) JobOption {
	return func(opts *JobOptions) error {
		if err := validateRetries(retries); err != nil {
			return err
		}
		opts.retries = retries
		return nil
	}
}

// WithJobRequeuing enables or disables requeuing if
// job task fails (default: true)
func WithJobRequeuing(enabled bool) JobOption {
	return func(opts *JobOptions) error {
		opts.requeuing = enabled
		return nil
	}
}

// WithJobWorkerPoolSize sets how many workers should
// handle this job(default: 1)
func WithJobWorkerPoolSize(size int) JobOption {
	return func(opts *JobOptions) error {
		if err := validatePoolSize(size); err != nil {
			return err
		}
		opts.workerPoolSize = size
		return nil
	}
}

// TaskOptions contains all task options
type TaskOptions struct {
	startAt        time.Time
	startAtEnabled bool
	retries        int
}

var defaultTaskOptions = TaskOptions{
	startAtEnabled: false,
	retries:        5,
}

// TaskOption configres task
type TaskOption func(*TaskOptions) error

func (opts TaskOptions) with(args ...TaskOption) (TaskOptions, error) {
	for _, opt := range args {
		if err := opt(&opts); err != nil {
			return opts, err
		}
	}
	return opts, nil
}

// WithTaskStartTime enables and sets time when task should be executed (default: disabled)
func WithTaskStartTime(t time.Time) TaskOption {
	return func(opts *TaskOptions) error {
		if err := validateStartTime(t); err != nil {
			return err
		}
		opts.startAt = t
		opts.startAtEnabled = true
		return nil
	}
}

// WithTaskStartTimeDisabled disables task start time
func WithTaskStartTimeDisabled() TaskOption {
	return func(opts *TaskOptions) error {
		opts.startAtEnabled = false
		return nil
	}
}

// WithTaskRetries sets initial retry number for failed tasks (default: 5)
func WithTaskRetries(retries int) TaskOption {
	return func(opts *TaskOptions) error {
		if err := validateRetries(retries); err != nil {
			return err
		}
		opts.retries = retries
		return nil
	}
}
