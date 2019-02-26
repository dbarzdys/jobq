package jobq

import (
	"context"
	"errors"
	"testing"
	"time"
)

func Test_worker_isWorking(t *testing.T) {
	type fields struct {
		working bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "working",
			fields: fields{
				working: true,
			},
			want: true,
		},
		{
			name: "not_working",
			fields: fields{
				working: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &worker{
				working: tt.fields.working,
			}
			if got := w.IsWorking(); got != tt.want {
				t.Errorf("worker.isWorking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_worker_resume(t *testing.T) {
	type fields struct {
		working bool
		runch   chan bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "working",
			fields: fields{
				working: true,
			},
		},
		{
			name: "not_working",
			fields: fields{
				working: false,
				runch:   make(chan bool),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fields.runch != nil {
				go func(ch <-chan bool) {
					<-ch
				}(tt.fields.runch)
			}
			w := &worker{
				working: tt.fields.working,
				runch:   tt.fields.runch,
			}
			w.Resume()
		})
	}
}

func Test_worker_stop(t *testing.T) {
	wrk := &worker{
		working: true,
		stopch:  make(chan bool),
	}
	go func(ch chan bool) {
		<-ch
		ch <- true
	}(wrk.stopch)
	wrk.Stop()
	if wrk.working {
		t.Error("worker.stop(); worker did not stop")
	}
}

func Test_worker_pause(t *testing.T) {
	wrk := &worker{
		working: true,
	}
	wrk.Pause()
	if wrk.working {
		t.Error("worker.pause(); worker did not pause")
	}
}

func Test_worker_isStopping(t *testing.T) {
	type fields struct {
		working bool
		runch   chan bool
		stopch  chan bool
	}
	tests := []struct {
		name   string
		fields fields
		before func(fields)
		want   bool
	}{
		{
			name: "working",
			fields: fields{
				working: true,
			},
			want: false,
		},
		{
			name: "receive_runch",
			fields: fields{
				working: false,
				runch:   make(chan bool),
				stopch:  make(chan bool),
			},
			before: func(f fields) {
				f.runch <- true
			},
			want: false,
		},
		{
			name: "receive_stopch",
			fields: fields{
				working: false,
				runch:   make(chan bool),
				stopch:  make(chan bool),
			},
			before: func(f fields) {
				f.stopch <- true
				<-f.stopch
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &worker{
				working: tt.fields.working,
				runch:   tt.fields.runch,
				stopch:  tt.fields.stopch,
			}
			if tt.before != nil {
				go tt.before(tt.fields)
			}
			if got := w.isStopping(); got != tt.want {
				t.Errorf("worker.isStopping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_worker_start(t *testing.T) {
	type fields struct {
		working bool
		runch   chan bool
		stopch  chan bool
		store   Store
		jobName string
		job     Job
	}
	tests := []struct {
		name   string
		fields fields
		before func(*worker)
	}{
		{
			name: "receive_stopch",
			fields: fields{
				working: false,
				runch:   make(chan bool),
				stopch:  make(chan bool),
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						return nil
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return &mockTaskAction{
							taskRow: &TaskRow{
								id:      1,
								body:    nil,
								jobName: "test",
								retries: 5,
								startAt: nullTime{
									Valid: false,
								},
								timeout: nullTime{
									Valid: false,
								},
							},
						}, nil
					},
				},
			},
			before: func(w *worker) {
				<-w.runch
				w.Stop()
			},
		},
		{
			name: "when_working",
			fields: fields{
				working: true,
				runch:   make(chan bool),
				stopch:  make(chan bool),
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						return nil
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return &mockTaskAction{
							taskRow: &TaskRow{
								id:      1,
								body:    nil,
								jobName: "test",
								retries: 5,
								startAt: nullTime{
									Valid: false,
								},
								timeout: nullTime{
									Valid: false,
								},
							},
						}, nil
					},
				},
			},
			before: func(w *worker) {
				w.Stop()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &worker{
				working: tt.fields.working,
				runch:   tt.fields.runch,
				stopch:  tt.fields.stopch,
				store:   tt.fields.store,
				jobName: tt.fields.jobName,
				job:     tt.fields.job,
			}
			if tt.before != nil {
				go tt.before(w)
			}
			w.Start()
		})
	}
}

func Test_worker_work(t *testing.T) {
	type fields struct {
		store   Store
		jobName string
		job     Job
		working bool
		opts    JobOptions
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						return nil
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return &mockTaskAction{
							taskRow: &TaskRow{
								id:      1,
								body:    nil,
								jobName: "test",
								retries: 5,
								startAt: nullTime{
									Valid: false,
								},
								timeout: nullTime{
									Valid: false,
								},
							},
						}, nil
					},
				},
				opts: defaultJobOptions,
			},
			wantErr: false,
		},
		{
			name: "failed_dequeue",
			fields: fields{
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						return nil
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return nil, errors.New("test err")
					},
				},
				opts: defaultJobOptions,
			},
			wantErr: true,
		},
		{
			name: "no_rows",
			fields: fields{
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						return nil
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return nil, ErrEmptyQueue
					},
				},
				opts: defaultJobOptions,
			},
			wantErr: true,
		},
		{
			name: "expires",
			fields: fields{
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						time.Sleep(time.Millisecond * 101)
						return nil
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return &mockTaskAction{
							taskRow: &TaskRow{
								id:      1,
								body:    nil,
								jobName: "test",
								retries: 5,
								startAt: nullTime{
									Valid: false,
								},
								timeout: nullTime{
									Valid: false,
								},
							},
						}, nil
					},
				},
				opts: JobOptions{
					ttl: time.Millisecond * 100,
				},
			},
			wantErr: true,
		},
		{
			name: "requeue_with_retries",
			fields: fields{
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						return errors.New("test err")
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return &mockTaskAction{
							taskRow: &TaskRow{
								id:      1,
								body:    nil,
								jobName: "test",
								retries: 5,
								startAt: nullTime{
									Valid: false,
								},
								timeout: nullTime{
									Valid: false,
								},
							},
						}, nil
					},
				},
				opts: JobOptions{
					ttl:       time.Millisecond * 100,
					requeuing: true,
				},
			},
			wantErr: false,
		},
		{
			name: "requeue_with_timeout",
			fields: fields{
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						return errors.New("test err")
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return &mockTaskAction{
							taskRow: &TaskRow{
								id:      1,
								body:    nil,
								jobName: "test",
								retries: 0,
								startAt: nullTime{
									Valid: false,
								},
								timeout: nullTime{
									Valid: false,
								},
							},
						}, nil
					},
				},
				opts: JobOptions{
					ttl:            time.Millisecond * 100,
					requeuing:      true,
					timeout:        time.Minute,
					timeoutEnabled: true,
					retries:        5,
				},
			},
			wantErr: false,
		},
		{
			name: "requeue_fails",
			fields: fields{
				jobName: "test",
				job: &mockJob{
					onHandleTask: func(context.Context, *Task) error {
						return errors.New("test err")
					},
				},
				store: &mockStore{
					onDequeue: func(name string) (TaskAction, error) {
						return &mockTaskAction{
							taskRow: &TaskRow{
								id:      1,
								body:    nil,
								jobName: "test",
								retries: 5,
								startAt: nullTime{
									Valid: false,
								},
								timeout: nullTime{
									Valid: false,
								},
							},
							errRequeue: errors.New("requeue test err"),
						}, nil
					},
				},
				opts: JobOptions{
					ttl:       time.Millisecond * 100,
					requeuing: true,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &worker{
				store:   tt.fields.store,
				jobName: tt.fields.jobName,
				job:     tt.fields.job,
				working: tt.fields.working,
				opts:    tt.fields.opts,
			}
			if err := w.work(); (err != nil) != tt.wantErr {
				t.Errorf("worker.work() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_worker_handleWorkErr(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		working     bool
		wantWorking bool
	}{
		{
			name:        "no_err",
			working:     true,
			err:         nil,
			wantWorking: true,
		},
		{
			name:        "empty_queue",
			working:     true,
			err:         ErrEmptyQueue,
			wantWorking: false,
		},
		{
			name:        "canceled",
			working:     true,
			err:         ErrWorkCanceled,
			wantWorking: true,
		},
		{
			name:        "unhandled",
			working:     true,
			err:         errors.New("unhandled err"),
			wantWorking: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &worker{
				working: tt.working,
			}
			w.handleWorkErr(tt.err)
			if w.working != tt.wantWorking {
				t.Errorf("worker.handleWorkErr(); working = %v, want = %v", w.working, tt.wantWorking)
			}
		})
	}
}
