package jobq

import (
	"database/sql"
	"sync"
)

type workerPool struct {
	db             *sql.DB
	jobName        string
	job            Job
	opts           JobOptions
	workers        []*worker
	awaitingResume bool
	sync.RWMutex
}

func makeWorkerPool(db *sql.DB, jobName string, job Job, opts JobOptions) *workerPool {
	return &workerPool{
		db:      db,
		jobName: jobName,
		job:     job,
		opts:    opts,
		workers: []*worker{},
	}
}

func (wp *workerPool) resumeOne() {
	wp.RLock()
	awaiting := wp.awaitingResume
	workers := wp.workers
	wp.RUnlock()
	if awaiting {
		return
	}
	for _, w := range workers {
		if !w.isWorking() {
			wp.Lock()
			wp.awaitingResume = true
			wp.Unlock()
			go func() {
				w.start()
				wp.Lock()
				wp.awaitingResume = false
				wp.Unlock()
			}()
			return
		}
	}
}

func (wp *workerPool) fill() {
	wp.RLock()
	l := len(wp.workers)
	size := int(wp.opts.workerPoolSize)
	wp.RUnlock()
	for i := l; i < size; i++ {
		wp.add()
	}
}
func (wp *workerPool) start() {
	wp.fill()
	wp.RLock()
	for _, w := range wp.workers {
		if !w.isWorking() {
			w.start()
		}
	}
	wp.RUnlock()
}
func (wp *workerPool) stop() {
	wp.RLock()
	for _, w := range wp.workers {
		w.stop()
	}
	wp.RUnlock()
}

func (wp *workerPool) add() {
	wp.Lock()
	w := makeWorker(wp.db, wp.jobName, wp.job, wp.opts)
	go w.run()
	wp.workers = append(wp.workers, w)
	wp.Unlock()
}
