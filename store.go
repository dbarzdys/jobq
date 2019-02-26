package jobq

import (
	"database/sql"
	"errors"
)

var (
	ErrEmptyQueue = errors.New("queue is empty")
)

type TaskRow struct {
	id      int64
	jobName string
	body    []byte
	retries int
	timeout nullTime
	startAt nullTime
}

type TaskAction interface {
	Commit() error
	Rollback() error
	Requeue(*TaskRow) error
	Row() *TaskRow
}

type taskAction struct {
	tx Tx
	r  *TaskRow
}

func (act taskAction) Commit() error {
	return act.tx.Commit()
}

func (act taskAction) Rollback() error {
	return act.tx.Rollback()
}

func (act taskAction) Requeue(row *TaskRow) error {
	return requeueTask(act.tx, row)
}

func (act taskAction) Row() *TaskRow {
	return act.r
}

type Store interface {
	Dequeue(name string) (TaskAction, error)
	Queue(row *TaskRow) error
}

type store struct {
	db DB
}

func (s store) Dequeue(name string) (TaskAction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	row, err := dequeueTask(tx, name)
	if err == sql.ErrNoRows {
		tx.Rollback()
		return nil, ErrEmptyQueue
	} else if err != nil {
		tx.Rollback()
		return nil, err
	}
	return &taskAction{
		tx: tx,
		r:  row,
	}, nil
}

func (s store) Queue(row *TaskRow) error {
	return queueTask(s.db, row)
}
