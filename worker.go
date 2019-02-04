package jobq

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"
)

type worker struct {
	id      int
	db      *sql.DB
	jobName string
	job     Job
	working bool
	runch   chan bool
	stopch  chan bool
	opts    JobOptions
	sync.RWMutex
}

func makeWorker(db *sql.DB, jobName string, job Job, opts JobOptions) *worker {
	return &worker{
		db:      db,
		jobName: jobName,
		job:     job,
		working: false,
		runch:   make(chan bool),
		stopch:  make(chan bool),
	}
}
func (w *worker) isWorking() bool {
	w.RLock()
	working := w.working
	w.RUnlock()
	return working
}

func (w *worker) start() {
	if w.isWorking() {
		return
	}
	w.Lock()
	w.working = true
	w.Unlock()
	w.runch <- true
}

func (w *worker) stop() {
	w.pause()
	w.stopch <- true
	<-w.stopch
}

func (w *worker) pause() {
	w.Lock()
	w.working = false
	w.Unlock()
}

func (w *worker) isStopping() bool {
	for !w.isWorking() {
		select {
		case <-w.runch:
			return false
		case <-w.stopch:
			w.stopch <- true
			w.stopch = nil
			return true
		}
	}
	return false
}

func (w *worker) work() error {
	tx, err := w.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	row := new(taskRow)
	err = row.dequeue(w.jobName, tx)
	if err == sql.ErrNoRows {
		// ran out of work
		// TODO: add logs
		w.pause()
		time.Sleep(time.Millisecond * 100)
		return nil
	} else if err != nil {
		// some other error
		time.Sleep(time.Second)
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	task := &Task{row, true}
	err = w.job.HandleTask(ctx, task)
	if ctx.Err() != nil {
		return errors.New("canceled") // TODO: define error for this
	}
	if err != nil && w.opts.requeuing {
		if task.row.retries > 0 {
			task.row.retries--
		} else {
			task.row.retries = w.opts.retries
			task.row.timeout = nullTime{
				Valid: w.opts.timeoutEnabled,
				Time:  time.Now().Add(w.opts.timeout).UTC(),
			}
		}
		err = task.row.requeue(tx)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (w *worker) run() {
	for {
		if !w.isStopping() {
			err := w.work()
			_ = err
		}
		// TODO: do something with err
	}
}
