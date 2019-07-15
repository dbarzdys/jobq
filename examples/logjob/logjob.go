package logjob

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dbarzdys/jobq"
)

// Name of a job
const Name = "logjob"

// LogJob implements jobq.Job
type LogJob struct{}

// New creates a new job
func New() *LogJob {
	return new(LogJob)
}

// HandleTask prints log
func (*LogJob) HandleTask(ctx context.Context, tsk *jobq.Task) error {
	body := new(TaskBody)
	err := tsk.ScanBody(body)
	if err != nil {
		return err
	}
	fmt.Printf("workerID: %02d, taskUID: %s, taskID: %d, message: %s\n", tsk.WorkerID(), tsk.UID(), tsk.ID(), body.Message)
	return nil
}

// TaskBody of a LogJob
type TaskBody struct {
	Message string `json:"message"`
}

// Value returns encoded TaskBody
func (tb *TaskBody) Value() ([]byte, error) {
	return json.Marshal(tb)
}

// Scan decodes val and updates TaskBody
func (tb *TaskBody) Scan(val []byte) error {
	return json.Unmarshal(val, tb)
}
