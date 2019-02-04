package jobq

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type event struct {
	JobName string   `json:"job_name"`
	Timeout nullTime `json:"timeout"`
	StartAt nullTime `json:"start_at"`
}

type listenerOpts struct {
	minReconnectInterval time.Duration
	maxReconnectInterval time.Duration
	aliveCheckInterval   time.Duration
	callback             pq.EventCallbackType
}

type listener struct {
	events   chan *event
	conninfo string
	listenerOpts
	dbListener *pq.Listener
}

func (l *listener) connect() error {
	l.dbListener = pq.NewListener(l.conninfo,
		l.minReconnectInterval,
		l.maxReconnectInterval,
		l.callback,
	)
	return l.dbListener.Listen("jobq_task_created")
}
func (l *listener) listen() error {
	err := l.connect()
	if err != nil {
		time.Sleep(time.Second)
		return l.listen()
	}
	for {
		select {
		case ev, ok := <-l.dbListener.Notify:
			if !ok {
				return l.listen()
			}
			if ev == nil {
				continue
			}
			body := []byte(ev.Extra)
			e := new(event)
			json.Unmarshal(body, e)
			l.events <- e
		case <-time.After(l.aliveCheckInterval):
			err = l.dbListener.Ping()
			if err != nil {
				return l.listen()
			}
		}
	}
}

func makeListener(conninfo string, opts listenerOpts) *listener {
	return &listener{
		events:       make(chan *event),
		listenerOpts: opts,
		conninfo:     conninfo,
	}
}
