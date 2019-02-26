package jobq

import (
	"errors"
	"testing"
	"time"
)

type mockScanner struct {
	onScan func([]byte) error
}

func (s mockScanner) Scan(val []byte) error {
	return s.onScan(val)
}

type mockValuer struct {
	onValue func() ([]byte, error)
}

func (v mockValuer) Value() ([]byte, error) {
	return v.onValue()
}

func TestPreparedTask_Queue(t *testing.T) {
	type fields struct {
		jobName string
		body    Valuer
		options TaskOptions
	}
	type args struct {
		execer DBExecer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				jobName: "test",
				body: &mockValuer{
					onValue: func() ([]byte, error) {
						return nil, nil
					},
				},
				options: TaskOptions{
					retries:        5,
					startAtEnabled: false,
				},
			},
			args: args{
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
			wantErr: false,
		},
		{
			name: "with_db_err",
			fields: fields{
				jobName: "test",
				body: &mockValuer{
					onValue: func() ([]byte, error) {
						return nil, nil
					},
				},
				options: TaskOptions{
					retries:        5,
					startAtEnabled: false,
				},
			},
			args: args{
				execer: &mockDBExecer{
					wantErr: true,
				},
			},
			wantErr: true,
		},
		{
			name: "with_valuer_err",
			fields: fields{
				jobName: "test",
				body: &mockValuer{
					onValue: func() ([]byte, error) {
						return nil, errors.New("test err")
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := &PreparedTask{
				jobName: tt.fields.jobName,
				body:    tt.fields.body,
				options: tt.fields.options,
			}
			if err := pt.Queue(tt.args.execer); (err != nil) != tt.wantErr {
				t.Errorf("PreparedTask.Queue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
