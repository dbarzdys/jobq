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
		body bytea NOT NULL,
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

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (nt *NullTime) UnmarshalJSON(b []byte) error {
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

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	str := nt.Time.Format(nullTimeLayout)
	str = fmt.Sprintf("\"%s\"", str)
	return []byte(str), nil
}

func (nt *NullTime) Scan(value interface{}) error {
	nt.Time, nt.Valid = value.(time.Time)
	return nil
}

func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

type DBExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}
type DBQueryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type taskRow struct {
	id      int64
	jobName string
	body    []byte
	retries uint
	timeout NullTime
	startAt NullTime
}

func (row *taskRow) queue(e DBExecer) error {
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

func (row *taskRow) requeue(e DBExecer) error {
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
		row.jobName,
		row.body,
		row.retries,
		row.timeout,
		row.startAt,
	)
	return err
}

func (row *taskRow) dequeue(jobName string, q DBQueryer) error {
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
	rows, err := q.Query(stmt, jobName)
	if err != nil {
		return err
	}
	if !rows.Next() {
		return sql.ErrNoRows
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
		return err
	}
	row.jobName = jobName
	return nil
}
