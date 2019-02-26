package jobq

// PreparedTask contains details required for work
// and is used for creating task using DBExecer
type PreparedTask struct {
	jobName string
	body    Valuer
	options TaskOptions
}

// Task contains details required for work
// and is used for for Job handle function
type Task struct {
	row      *TaskRow
	requeue  bool
	workerID int
}

// ScanBody scans tasks row body with TaskBody implementation
func (tsk *Task) ScanBody(body Scanner) error {
	return body.Scan(tsk.row.body)
}

// ID returns unique task identifier
func (tsk *Task) ID() int64 {
	return tsk.row.id
}

// WorkerID returns worker identifier
func (tsk *Task) WorkerID() int {
	return tsk.workerID
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
func NewTask(jobName string, body Valuer, opts ...TaskOption) (*PreparedTask, error) {
	if err := firstError(
		validateJobName(jobName),
		validateTaskBodyValuer(body),
	); err != nil {
		return nil, err
	}
	options, err := defaultTaskOptions.with(opts...)
	if err != nil {
		return nil, err
	}
	task := PreparedTask{
		jobName: jobName,
		body:    body,
		options: options,
	}
	return &task, nil
}

func (pt *PreparedTask) row() (*TaskRow, error) {
	body, err := pt.body.Value()
	if err != nil {
		return nil, err
	}
	return &TaskRow{
		jobName: pt.jobName,
		body:    body,
		retries: pt.options.retries,
		startAt: nullTime{
			Valid: pt.options.startAtEnabled,
			Time:  pt.options.startAt.UTC(),
		},
	}, nil
}

// Queue pushes PreparedTask to task queue
func (pt *PreparedTask) Queue(e DBExecer) error {
	row, err := pt.row()
	if err != nil {
		return err
	}
	return queueTask(e, row)
}
