package jobq

import (
	"database/sql"
	"sync"
)

type workerPool struct {
	db      *sql.DB
	jobName string
	job     Job
	maxSize int
	workers []*worker
	sync.RWMutex
}

func makeWorkerPool(db *sql.DB, jobName string, job Job, max int) *workerPool {
	return &workerPool{
		db:      db,
		jobName: jobName,
		job:     job,
		maxSize: max,
		workers: []*worker{},
	}
}

func (wp *workerPool) resumeOne() {
	wp.RLock()
	for _, w := range wp.workers {
		if !w.isWorking() {
			w.start()
			wp.RUnlock()
			return
		}
	}
	wp.RUnlock()
}

func (wp *workerPool) fill() {
	wp.RLock()
	l := len(wp.workers)
	wp.RUnlock()
	for i := l; i < wp.maxSize; i++ {
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
	w := makeWorker(wp.db, wp.jobName, wp.job)
	go w.run()
	wp.workers = append(wp.workers, w)
	wp.Unlock()
}
