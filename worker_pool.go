package jobq

import (
	"sync"
)

type WorkerPool interface {
	Resume(n int)
	Scale(size int)
	Start()
	Stop()
}

type workerPool struct {
	workers  []Worker
	factory  WorkerFactory
	resuming bool
	sync.RWMutex
}

func NewWorkerPool(factory WorkerFactory) WorkerPool {
	return &workerPool{
		factory: factory,
		workers: []Worker{},
	}
}

func (wp *workerPool) Resume(n int) {
	wp.Lock()
	defer wp.Unlock()
	if wp.resuming {
		return
	}
	wp.resuming = true
	for _, w := range wp.workers {
		if n == 0 {
			break
		}
		if !w.IsWorking() {
			w.Resume()
			n--
		}
	}
	wp.resuming = false
}

func (wp *workerPool) Scale(size int) {
	l := len(wp.workers)
	for l != size {
		if l < size {
			wp.increase()
			l++
		} else {
			wp.decrease()
			l--
		}
	}
}

func (wp *workerPool) Start() {
	wp.RLock()
	workers := wp.workers
	wp.RUnlock()
	for _, w := range workers {
		if !w.IsWorking() {
			w.Resume()
		}
	}
}

func (wp *workerPool) Stop() {
	wp.RLock()
	defer wp.RUnlock()
	for _, w := range wp.workers {
		w.Stop()
	}
}

func (wp *workerPool) increase() {
	w := wp.factory.Make()
	go w.Start()
	wp.workers = append(wp.workers, w)
}

func (wp *workerPool) decrease() {
	l := len(wp.workers)
	if l == 0 {
		return
	}
	w := wp.workers[l-1]
	w.Stop()
	wp.workers = wp.workers[0 : l-1]
}
