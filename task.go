package jobq

import "errors"

// PreparedTask contains details required for work
// and is used for creating task using DBExecer
type PreparedTask struct {
	jobName string
	body    Valuer
	opts    TaskOptions
}

// Task contains details required for work
// and is used for for Job handle function
type Task struct {
	row     *taskRow
	requeue bool
}

// ScanBody scans tasks row body with TaskBody implementation
func (tsk *Task) ScanBody(body Scanner) error {
	return body.Scan(tsk.row.body)
}

// ID returns unique task identifier
func (tsk *Task) ID() int64 {
	return tsk.row.id
}

// TaskBody scans and returns value of a task body using []byte
type TaskBody interface {
	Scanner
	Valuer
}

// Scanner scans value using []byte
type Scanner interface {
	Scan(val []byte) error
}

// Valuer returns value using []byte
type Valuer interface {
	Value() ([]byte, error)
}

// NewTask creates a new PreparedTask
func NewTask(jobName string, body Valuer, options ...TaskOption) *PreparedTask {
	o := defaultTaskOptions
	for _, opt := range options {
		opt(&o)
	}
	return &PreparedTask{
		jobName: jobName,
		body:    body,
		opts:    o,
	}
}

func (pt *PreparedTask) validate() error {
	if pt == nil {
		return errors.New("job not defined")
	}
	if pt.body == nil {
		return errors.New("queue not defined")
	}
	return nil
}

func (pt *PreparedTask) row() (*taskRow, error) {
	err := pt.validate()
	if err != nil {
		return nil, err
	}
	body, err := pt.body.Value()
	if err != nil {
		return nil, err
	}
	return &taskRow{
		jobName: pt.jobName,
		body:    body,
		retries: pt.opts.retries,
		startAt: nullTime{
			Valid: pt.opts.startAtEnabled,
			Time:  pt.opts.startAt.UTC(),
		},
	}, nil
}

// Queue pushes PreparedTask to task queue
func (pt *PreparedTask) Queue(e DBExecer) error {
	row, err := pt.row()
	if err != nil {
		return err
	}
	return row.queue(e)
}
