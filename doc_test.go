package jobq_test

import (
	"database/sql"
	"time"

	"github.com/dbarzdys/jobq"
	"github.com/dbarzdys/jobq/examples/logjob"
)

// This example shows how to create and run
// your job manager
func ExampleManager() {
	var conninfo string
	manager := jobq.NewManager(conninfo)
	manager.Run()
}

// This example shows how to create a new job
// and register it using job manager
func ExampleJob() {
	var manager *jobq.Manager
	opts := []jobq.JobOption{
		jobq.WithJobWorkerPoolSize(1),
		jobq.WithJobRequeuing(true),
		jobq.WithJobRequeueRetries(5),
		jobq.WithJobTimeout(time.Second * 5),
	}
	err := manager.Register(logjob.Name, logjob.New(), opts...)
	if err != nil {
		return
	}
}

// This example shows how to create
// new prepared task
func ExampleNewTask() {
	var tx *sql.Tx
	body := &logjob.TaskBody{
		Message: "Hello World",
	}
	opts := []jobq.TaskOption{
		jobq.WithTaskRetries(5),
		jobq.WithTaskStartTime(time.Now().Add(time.Hour * 24)),
	}
	task, err := jobq.NewTask(logjob.Name, body, opts...)
	if err != nil {
		return
	}
	err = task.Queue(tx)
	if err != nil {
		return
	}
}

// This example shows how to save
// prepared task to database using transaction
func ExamplePreparedTask() {
	var db *sql.DB
	var task *jobq.PreparedTask

	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	// create or update other tables

	task.Queue(tx)
	tx.Commit()
}
