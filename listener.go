package jobq

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

type event struct {
	JobName string   `json:"job_name"`
	Timeout NullTime `json:"timeout"`
	StartAt NullTime `json:"start_at"`
}

type listenerOpts struct {
	minReconnectInterval time.Duration
	maxReconnectInterval time.Duration
	aliveCheckInterval   time.Duration
	callback             pq.EventCallbackType
}

type listener struct {
	events chan *event
	listenerOpts
	dbListener *pq.Listener
}

func (l *listener) listen() error {
	err := l.dbListener.Listen("jobq_task_created")
	if err != nil {
		return err
	}
	for {
		select {
		case ev := <-l.dbListener.Notify:
			body := []byte(ev.Extra)
			e := new(event)
			json.Unmarshal(body, e)
			l.events <- e
		case <-time.After(l.aliveCheckInterval):
			go l.dbListener.Ping()
		}
	}
}

func makeListener(conninfo string, opts listenerOpts) *listener {
	return &listener{
		events:       make(chan *event),
		listenerOpts: opts,
		dbListener: pq.NewListener(conninfo,
			opts.minReconnectInterval,
			opts.maxReconnectInterval,
			opts.callback,
		),
	}
}
