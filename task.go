package jobq

import "errors"

type UnqueuedTask struct {
	jobName string
	body    TaskBody
	opts    TaskOptions
}

type Task struct {
	row     *taskRow
	requeue bool
}

func (tsk *Task) ScanBody(body TaskBody) error {
	return body.Scan(tsk.row.body)
}

func (tsk *Task) ID() int64 {
	return tsk.row.id
}

type TaskBody interface {
	Value() ([]byte, error)
	Scan(val []byte) error
}

func NewTask(jobName string, body TaskBody, options ...TaskOption) *UnqueuedTask {
	o := DefaultTaskOptions
	for _, opt := range options {
		opt(&o)
	}
	return &UnqueuedTask{
		jobName: jobName,
		body:    body,
		opts:    o,
	}
}

func (js *UnqueuedTask) validate() error {
	if js == nil {
		return errors.New("job not defined")
	}
	if js.body == nil {
		return errors.New("queue not defined")
	}
	return nil
}

func (js *UnqueuedTask) row() (*taskRow, error) {
	err := js.validate()
	if err != nil {
		return nil, err
	}
	body, err := js.body.Value()
	if err != nil {
		return nil, err
	}
	return &taskRow{
		jobName: js.jobName,
		body:    body,
		retries: js.opts.retries,
		startAt: NullTime{
			Valid: js.opts.startAtEnabled,
			Time:  js.opts.startAt.UTC(),
		},
	}, nil
}

func (js *UnqueuedTask) Queue(e DBExecer) error {
	row, err := js.row()
	if err != nil {
		return err
	}
	return row.queue(e)
}
