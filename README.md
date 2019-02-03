# jobq

Transactional job queue using PostgreSQL database.

## Goals

* Transactional job processing
* Concurrent processing
* Retries
* Scheduled jobs
* Multiple queues

## Example
1. You create your job
``` go
package logjob

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	fmt.Printf("at: %v, taskID: %d, message: %s\n", time.Now(), tsk.ID(), body.Message)
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
```
2. Create job manager
``` go
...
    manager := jobq.NewManager(conninfo)
...

```
3. Register your job
``` go
...
    manager.Register(logjob.Name, logjob.New())
...
```
3. Run job manager
``` go
...
    go manager.Run()
    defer manager.Close()
...
```

4. Create tasks
``` go
...
    db, err := sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
    }
    defer db.Close()
    task := jobq.NewTask(logjob.Name, &logjob.TaskBody{
        Message: "Hello World",
    })
    err = task.Queue(db)
...

```
