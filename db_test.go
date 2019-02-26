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

type mockTx struct {
	*mockDBExecer
	*mockDBQueryer
	onCommit   error
	onRollback error
}

func (tx mockTx) Commit() error {
	return tx.onCommit
}

func (tx mockTx) Rollback() error {
	return tx.onRollback
}

type mockDB struct {
	*mockDBExecer
	*mockDBQueryer
	onTx func() (*sql.Tx, error)
}

func (db mockDB) Begin() (*sql.Tx, error) {
	return db.onTx()
}

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
