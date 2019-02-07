package jobq

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Manager manages jobs and workers
type Manager struct {
	conninfo string
	db       *sql.DB
	listener *listener
	pools    map[string]*workerPool
	jobs     map[string]Job
	opts     map[string]JobOptions
	stopch   chan bool
}

// NewManager creates a new Manager using conninfo for database connection
func NewManager(conninfo string) *Manager {
	return &Manager{
		conninfo: conninfo,
		jobs:     make(map[string]Job),
		pools:    make(map[string]*workerPool),
		opts:     make(map[string]JobOptions),
	}
}

// Register adds a new job that will be handled by it's own workers
func (m *Manager) Register(name string, job Job, opts ...JobOption) error {
	if err := firstError(
		validateJobName(name),
		validateIfJobUnregistered(name, m.jobs),
	); err != nil {
		return err
	}
	options, err := defaultJobOptions.with(opts...)
	if err != nil {
		return err
	}
	m.opts[name] = options
	m.jobs[name] = job
	return nil
}

// Close stops all workers and closes connection to database
func (m *Manager) Close() (err error) {
	if m.stopch == nil {
		return
	}
	m.stopch <- true
	<-m.stopch
	return
}

// Run will connect to database and will start all workers
func (m *Manager) Run() (err error) {
	if err = m.setupDB(); err != nil {
		return err
	}
	m.setupWorkerPools()
	m.setupListener()
	// create stop channel
	m.stopch = make(chan bool)
	// create error channel
	errch := make(chan error)
	go func() {
		for _, p := range m.pools {
			p.start()
		}
	}()
	go func(ch chan<- error) {
		ch <- m.listener.listen()
	}(errch)
	for {
		select {
		// stop
		case <-m.stopch:
			for _, p := range m.pools {
				p.stop()
			}
			m.stopch <- true
			m.stopch = nil
			return
		// error
		case err = <-errch:
			go m.Close()
			return
		// event received
		case ev := <-m.listener.events:
			pool, ok := m.pools[ev.JobName]
			if ok {
				pool.resumeOne()
			}
		case <-time.After(time.Second * 5):
			for _, p := range m.pools {
				p.resumeOne()
			}
		}
	}
}

func (m *Manager) setupListener() {
	m.listener = makeListener(m.conninfo, listenerOpts{
		aliveCheckInterval:   time.Second * 60,
		minReconnectInterval: 10 * time.Second,
		maxReconnectInterval: time.Minute,
		callback: func(ev pq.ListenerEventType, err error) {
			// TODO:
		},
	})
}
func (m *Manager) setupDB() error {
	db, err := sql.Open("postgres", m.conninfo)
	if err != nil {
		return err
	}
	_, err = db.Exec(dbSchema)
	if err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *Manager) setupWorkerPools() {
	for name, job := range m.jobs {
		opts := m.opts[name]
		m.pools[name] = makeWorkerPool(m.db, name, job, opts)
	}
}
