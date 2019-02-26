package jobq

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

const dbSchema = `
	CREATE TABLE IF NOT EXISTS jobq_tasks (
		id BIGSERIAL,
		job_name varchar(100) NOT NULL,
		body jsonb NOT NULL,
		retries int NOT NULL,
		timeout timestamp,
		start_at timestamp,
		PRIMARY KEY(id)
	);

	CREATE OR REPLACE FUNCTION jobq_notify_task_created() RETURNS TRIGGER AS $$
    DECLARE 
        notification jsonb;
    BEGIN
        notification = json_build_object(
			'job_name', NEW.job_name,
			'timeout', NEW.timeout,
			'start_at', NEW.start_at
		);
        PERFORM pg_notify('jobq_task_created', notification::text);
        RETURN NULL; 
    END;
	$$ LANGUAGE plpgsql;

	
	DO $$ BEGIN
		IF NOT EXISTS(SELECT *
			FROM information_schema.triggers
			WHERE event_object_table = 'jobq_tasks'
			AND trigger_name = 'jobq_task_trigger'
			)
			THEN
				CREATE TRIGGER jobq_task_trigger
					AFTER INSERT ON jobq_tasks
					FOR EACH ROW EXECUTE PROCEDURE jobq_notify_task_created();
			
			END IF ;
		END;
	$$
	`

const nullTimeLayout = "2006-01-02T15:04:05.999999999"

type nullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *nullTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		nt.Valid = false
		return nil
	}
	str := string(b)
	str = str[1 : len(str)-1]
	t, err := time.Parse(nullTimeLayout, str)
	if err != nil {
		return err
	}
	nt.Time = t
	nt.Valid = true
	return nil
}

func (nt nullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	str := nt.Time.Format(nullTimeLayout)
	str = fmt.Sprintf("\"%s\"", str)
	return []byte(str), nil
}

func (nt *nullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

func (nt nullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// DBExecer makes execs
type DBExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// DBQueryer makes queries
type DBQueryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type DB interface {
	DBExecer
	DBQueryer
	Begin() (*sql.Tx, error)
}

type Tx interface {
	DBExecer
	DBQueryer
	Commit() error
	Rollback() error
}

func queueTask(e DBExecer, row *TaskRow) error {
	stmt := `
		INSERT INTO jobq_tasks (
			job_name,
			body,
			retries,
			timeout,
			start_at
		) VALUES ($1, $2, $3, $4, $5);
	`
	_, err := e.Exec(
		stmt,
		row.jobName,
		row.body,
		row.retries,
		row.timeout,
		row.startAt,
	)
	return err
}

func requeueTask(e DBExecer, row *TaskRow) error {
	stmt := `
		INSERT INTO jobq_tasks (
			id,
			job_name,
			body,
			retries,
			timeout,
			start_at
		) VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := e.Exec(
		stmt,
		row.id,
		row.jobName,
		row.body,
		row.retries,
		row.timeout,
		row.startAt,
	)
	return err
}

func dequeueTask(e DBQueryer, name string) (*TaskRow, error) {
	row := new(TaskRow)
	stmt := `
		DELETE FROM jobq_tasks WHERE id = (
			SELECT id FROM jobq_tasks
			WHERE job_name = $1
			AND (timeout IS NULL OR timeout < NOW())
			AND (start_at IS NULL OR start_at < NOW())
			ORDER BY id ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		) RETURNING id, body, retries, timeout, start_at;
	`
	rows, err := e.Query(stmt, name)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	defer rows.Close()
	err = rows.Scan(
		&row.id,
		&row.body,
		&row.retries,
		&row.timeout,
		&row.startAt,
	)
	if err != nil {
		return nil, err
	}
	row.jobName = name
	return row, nil
}
