package jobq

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrEmptyQueue = errors.New("queue is empty")
)

func uuid() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	var u [16]byte
	copy(u[:], buf[:])
	u[8] = (u[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

type TaskRow struct {
	id      int64
	uid     string
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
