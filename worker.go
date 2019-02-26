package jobq

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrWorkCanceled = errors.New("work has been canceled")
)

type Worker interface {
	ID() int
	IsWorking() bool
	Start()
	Stop()
	Resume()
	Pause()
}

type worker struct {
	id       int
	store    Store
	jobName  string
	job      Job
	working  bool
	awaiting bool
	runch    chan bool
	okch     chan bool
	stopch   chan bool
	opts     JobOptions
	sync.RWMutex
}

type WorkerFactory interface {
	WithJob(name string, job Job) WorkerFactory
	WithStore(store Store) WorkerFactory
	WithOptions(opts JobOptions) WorkerFactory
	Make() Worker
}

type workerFactory struct {
	n       int
	jobName string
	job     Job
	store   Store
	opts    JobOptions
}

func (f *workerFactory) WithJob(name string, job Job) WorkerFactory {
	f.jobName = name
	f.job = job
	return f
}

func (f *workerFactory) WithStore(store Store) WorkerFactory {
	f.store = store
	return f
}

func (f *workerFactory) WithOptions(opts JobOptions) WorkerFactory {
	f.opts = opts
	return f
}

func NewWorkerFactory() WorkerFactory {
	return &workerFactory{}
}

func (f *workerFactory) Make() Worker {
	f.n++
	return &worker{
		id:      f.n,
		jobName: f.jobName,
		job:     f.job,
		store:   f.store,
		opts:    f.opts,
		working: false,
		runch:   make(chan bool),
		okch:    make(chan bool),
		stopch:  make(chan bool),
	}
}

func (w *worker) ID() int {
	return w.id
}

func (w *worker) IsWorking() bool {
	w.RLock()
	defer w.RUnlock()
	return w.working
}

func (w *worker) Start() {
	for !w.isStopping() {
		err := w.work()
		w.handleWorkErr(err)
	}
}

func (w *worker) handleWorkErr(err error) {
	switch err {
	case nil:
		return
	case ErrEmptyQueue:
		w.Pause()
		time.Sleep(time.Millisecond * 100)
		return
	case ErrWorkCanceled:
		time.Sleep(time.Second)
		return
	default:
		fmt.Printf("unhandled err: %v\n", err)
		time.Sleep(time.Second)
		return
	}
}

func (w *worker) Stop() {
	w.Pause()
	w.stopch <- true
	<-w.stopch
}

func (w *worker) Resume() {
	w.runch <- true
	<-w.okch
	w.Lock()
	w.working = true
	w.Unlock()
}

func (w *worker) Pause() {
	w.Lock()
	defer w.Unlock()
	w.working = false
}

func (w *worker) isStopping() bool {
	for !w.IsWorking() {
		select {
		case <-w.runch:
			w.okch <- true
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
	act, err := w.store.Dequeue(w.jobName)
	if err != nil {
		return err
	}
	defer act.Rollback()
	row := act.Row()
	ctx, cancel := context.WithTimeout(context.Background(), w.opts.ttl)
	defer cancel()
	task := &Task{row, true, w.id}
	err = w.job.HandleTask(ctx, task)
	if err != nil && w.opts.requeuing {
		prepareTaskForRequeue(task, w.opts)
		err = act.Requeue(row)
		if err != nil {
			return err
		}
	}
	if err = ctx.Err(); err != nil {
		return ErrWorkCanceled
	}

	return act.Commit()
}

func prepareTaskForRequeue(task *Task, opts JobOptions) {
	if task.row.retries > 0 {
		task.row.retries--
	} else {
		task.row.retries = opts.retries
		task.row.timeout = nullTime{
			Valid: opts.timeoutEnabled,
			Time:  time.Now().Add(opts.timeout).UTC(),
		}
	}
}
