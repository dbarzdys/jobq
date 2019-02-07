package logjob

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dbarzdys/jobq"
)

const Name = "logjob"

type LogJob struct{}

func New() *LogJob {
	return new(LogJob)
}

func (*LogJob) HandleTask(ctx context.Context, tsk *jobq.Task) error {
	body := new(TaskBody)
	err := tsk.ScanBody(body)
	if err != nil {
		return err
	}
	fmt.Printf("taskID: %d, message: %s\n", tsk.ID(), body.Message)
	return nil
}

type TaskBody struct {
	Message string `json:"message"`
}

func (tb *TaskBody) Value() ([]byte, error) {
	return json.Marshal(tb)
}

func (tb *TaskBody) Scan(val []byte) error {
	return json.Unmarshal(val, tb)
}
