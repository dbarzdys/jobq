package jobq

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_nullTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []byte
		want    nullTime
		wantErr bool
	}{
		{
			name: "success_not_null",
			want: nullTime{
				Time:  time.Unix(10000, 1).UTC(),
				Valid: true,
			},
			bytes: []byte("\"1970-01-01T02:46:40.000000001\""),
		},
		{
			name: "success_null",
			want: nullTime{
				Valid: false,
			},
			bytes: []byte("null"),
		},
		{
			name:    "fail_null",
			bytes:   []byte("invalid string here"),
			wantErr: true,
		},
		{
			name:    "fail_not_null",
			bytes:   []byte("1970-01"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nt := &nullTime{}
			if err := nt.UnmarshalJSON(tt.bytes); (err != nil) != tt.wantErr {
				t.Errorf("nullTime.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if nt.Valid != tt.want.Valid {
				t.Errorf("nullTime.Valid = %v, want %v", nt.Valid, tt.want.Valid)
			}
			if !tt.want.Valid && nt.Time.Unix() != tt.want.Time.Unix() {
				t.Errorf("nullTime.Time = %v, want %v", nt.Time.Unix(), tt.want.Time.Unix())
			}
		})
	}
}

func Test_nullTime_MarshalJSON(t *testing.T) {
	type fields struct {
		Time  time.Time
		Valid bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "null",
			fields: fields{
				Valid: false,
			},
			want: []byte("null"),
		},
		{
			name: "not_null",
			fields: fields{
				Time:  time.Unix(10000, 1).UTC(),
				Valid: true,
			},
			want: []byte("\"1970-01-01T02:46:40.000000001\""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nt := nullTime{
				Time:  tt.fields.Time,
				Valid: tt.fields.Valid,
			}
			got, err := nt.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("nullTime.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nullTime.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nullTime_Scan(t *testing.T) {
	var nt nullTime
	nt.Scan(time.Unix(10000, 1))
	want := nullTime{
		Time:  time.Unix(10000, 1),
		Valid: true,
	}
	if !reflect.DeepEqual(nt, want) {
		t.Errorf("nullTime.Scan() = %v, want %v", nt, want)
	}
}

func Test_nullTime_Value(t *testing.T) {
	type fields struct {
		Time  time.Time
		Valid bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    driver.Value
		wantErr bool
	}{
		{
			name: "null",
			fields: fields{
				Valid: false,
			},
			want: nil,
		},
		{
			name: "not_null",
			fields: fields{
				Time:  time.Unix(10000, 1).UTC(),
				Valid: true,
			},
			want: time.Unix(10000, 1).UTC(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nt := nullTime{
				Time:  tt.fields.Time,
				Valid: tt.fields.Valid,
			}
			got, err := nt.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("nullTime.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nullTime.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockDBExecer struct {
	wantStmt string
	wantArgs []interface{}
	wantErr  bool
	gotStmt  string
	gotArgs  []interface{}
	valid    bool
}

func (e *mockDBExecer) Exec(stmt string, args ...interface{}) (sql.Result, error) {
	e.gotStmt = stmt
	e.gotArgs = args
	stmt = strings.Replace(stmt, "\n", "", -1)
	stmt = strings.Replace(stmt, "\t", "", -1)
	wantStmt := e.wantStmt
	wantStmt = strings.Replace(wantStmt, "\n", "", -1)
	wantStmt = strings.Replace(wantStmt, "\t", "", -1)
	if stmt != wantStmt {
		goto end
	}
	if len(args) != len(e.wantArgs) {
		goto end
	}
	if !reflect.DeepEqual(args, e.wantArgs) {
		goto end
	}
	e.valid = true
end:
	if e.wantErr {
		return nil, errors.New("mock err")
	}
	return nil, nil
}

type mockDBQueryer struct {
	wantStmt string
	wantArgs []interface{}
	wantErr  bool
	gotStmt  string
	gotArgs  []interface{}
	valid    bool
}

func (e *mockDBQueryer) Query(stmt string, args ...interface{}) (*sql.Rows, error) {
	e.gotStmt = stmt
	e.gotArgs = args
	stmt = strings.Replace(stmt, "\n", "", -1)
	stmt = strings.Replace(stmt, "\t", "", -1)
	wantStmt := e.wantStmt
	wantStmt = strings.Replace(wantStmt, "\n", "", -1)
	wantStmt = strings.Replace(wantStmt, "\t", "", -1)
	if stmt != wantStmt {
		goto end
	}
	if len(args) != len(e.wantArgs) {
		goto end
	}
	if !reflect.DeepEqual(args, e.wantArgs) {
		goto end
	}
	e.valid = true
end:
	if e.wantErr {
		return nil, errors.New("mock err")
	}
	return nil, nil
}

func Test_taskRow_queue(t *testing.T) {
	type fields struct {
		id      int64
		jobName string
		body    []byte
		retries uint
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
					uint(5),
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
			row := &taskRow{
				id:      tt.fields.id,
				jobName: tt.fields.jobName,
				body:    tt.fields.body,
				retries: tt.fields.retries,
				timeout: tt.fields.timeout,
				startAt: tt.fields.startAt,
			}
			if err := row.queue(tt.execer); (err != nil) != tt.wantErr {
				t.Errorf("taskRow.queue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.execer.valid {
				t.Errorf("taskRow.queue() gotStmt = %s, wantStmt = %s", tt.execer.gotStmt, tt.execer.wantStmt)
				t.Errorf("taskRow.queue() gotArgs = %v, wantArgs = %v", tt.execer.gotArgs, tt.execer.wantArgs)
			}
		})
	}
}

func Test_taskRow_requeue(t *testing.T) {
	type fields struct {
		id      int64
		jobName string
		body    []byte
		retries uint
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
					uint(5),
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
			row := &taskRow{
				id:      tt.fields.id,
				jobName: tt.fields.jobName,
				body:    tt.fields.body,
				retries: tt.fields.retries,
				timeout: tt.fields.timeout,
				startAt: tt.fields.startAt,
			}
			if err := row.requeue(tt.execer); (err != nil) != tt.wantErr {
				t.Errorf("taskRow.requeue() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.execer.valid {
				t.Errorf("taskRow.queue() gotStmt = %s, wantStmt = %s", tt.execer.gotStmt, tt.execer.wantStmt)
				t.Errorf("taskRow.queue() gotArgs = %v, wantArgs = %v", tt.execer.gotArgs, tt.execer.wantArgs)
			}
		})
	}
}
