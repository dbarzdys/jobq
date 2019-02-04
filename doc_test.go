package jobq_test

import (
	"database/sql"

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
	manager.Register(logjob.Name, logjob.New())
}

// This example shows how to create
// new prepared task
func ExampleNewTask() {
	body := &logjob.TaskBody{
		Message: "Hello World",
	}
	jobq.NewTask(logjob.Name, body)
}

// This example shows how to save
// prepared task to database
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
