package jobq

import (
	"testing"
	"time"
)

type mockTaskAction struct {
	errCommit   error
	errRollback error
	errRequeue  error
	taskRow     *TaskRow
}

func (act mockTaskAction) Commit() error {
	return act.errCommit
}
func (act mockTaskAction) Rollback() error {
	return act.errRollback
}
func (act mockTaskAction) Requeue(*TaskRow) error {
	return act.errRequeue

}
func (act mockTaskAction) Row() *TaskRow {
	return act.taskRow

}

type mockStore struct {
	onDequeue func(name string) (TaskAction, error)
	onQueue   func(row *TaskRow) error
}

func (store *mockStore) Dequeue(name string) (TaskAction, error) {
	return store.onDequeue(name)
}

func (store *mockStore) Queue(row *TaskRow) error {
	return store.onQueue(row)
}

func Test_storeImpl_queue(t *testing.T) {
	type fields struct {
		id      int64
		jobName string
		body    []byte
		retries int
		timeout nullTime
		startAt nullTime
	}
	tests := []struct {
		name    string
		fields  fields
		execer  *mockDBExecer
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				jobName: "test-job-name",
				body:    []byte{0, 1, 2, 3},
				retries: 5,
				timeout: nullTime{
					Valid: true,
					Time:  time.Unix(1000, 0),
				},
				startAt: nullTime{
					Valid: true,
					Time:  time.Unix(2000, 0),
				},
			},
			execer: &mockDBExecer{
				wantStmt: `
					INSERT INTO jobq_tasks (
						job_name,
						body,
						retries,
						timeout,
						start_at
					) VALUES ($1, $2, $3, $4, $5);
				`,
				wantArgs: []interface{}{
					"test-job-name",
					[]byte{0, 1, 2, 3},
					5,
					nullTime{
						Valid: true,
						Time:  time.Unix(1000, 0),
					},
					nullTime{
						Valid: true,
						Time:  time.Unix(2000, 0),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := &TaskRow{
				id:      tt.fields.id,
				jobName: tt.fields.jobName,
				body:    tt.fields.body,
				retries: tt.fields.retries,
				timeout: tt.fields.timeout,
				startAt: tt.fields.startAt,
			}
			store := &store{
				&mockDB{
					mockDBExecer: tt.execer,
				},
			}
			if err := store.Queue(row); (err != nil) != tt.wantErr {
				t.Errorf("storeImpl.queue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.execer.valid {
				t.Errorf("storeImpl.queue() gotStmt = %s, wantStmt = %s", tt.execer.gotStmt, tt.execer.wantStmt)
				t.Errorf("storeImpl.queue() gotArgs = %v, wantArgs = %v", tt.execer.gotArgs, tt.execer.wantArgs)
			}
		})
	}
}

func Test_taskActionImpl_requeue(t *testing.T) {
	type fields struct {
		id      int64
		jobName string
		body    []byte
		retries int
		timeout nullTime
		startAt nullTime
	}
	tests := []struct {
		name    string
		fields  fields
		execer  *mockDBExecer
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				id:      100,
				jobName: "test-job-name",
				body:    []byte{0, 1, 2, 3},
				retries: 5,
				timeout: nullTime{
					Valid: true,
					Time:  time.Unix(1000, 0),
				},
				startAt: nullTime{
					Valid: true,
					Time:  time.Unix(2000, 0),
				},
			},
			execer: &mockDBExecer{
				wantStmt: `
					INSERT INTO jobq_tasks (
						id,
						job_name,
						body,
						retries,
						timeout,
						start_at
					) VALUES ($1, $2, $3, $4, $5, $6);
				`,
				wantArgs: []interface{}{
					int64(100),
					"test-job-name",
					[]byte{0, 1, 2, 3},
					5,
					nullTime{
						Valid: true,
						Time:  time.Unix(1000, 0),
					},
					nullTime{
						Valid: true,
						Time:  time.Unix(2000, 0),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := &TaskRow{
				id:      tt.fields.id,
				jobName: tt.fields.jobName,
				body:    tt.fields.body,
				retries: tt.fields.retries,
				timeout: tt.fields.timeout,
				startAt: tt.fields.startAt,
			}
			act := &taskAction{
				tx: &mockTx{
					mockDBExecer: tt.execer,
				},
			}
			if err := act.Requeue(row); (err != nil) != tt.wantErr {
				t.Errorf("taskActionImpl.requeue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.execer.valid {
				t.Errorf("taskActionImpl.queue() gotStmt = %s, wantStmt = %s", tt.execer.gotStmt, tt.execer.wantStmt)
				t.Errorf("taskActionImpl.queue() gotArgs = %v, wantArgs = %v", tt.execer.gotArgs, tt.execer.wantArgs)
			}
		})
	}
}
