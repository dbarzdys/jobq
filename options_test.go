package jobq

import (
	"testing"
	"time"
)

func TestWithJobTimeout(t *testing.T) {
	tests := []struct {
		name               string
		opts               JobOptions
		timeout            time.Duration
		wantTimeout        time.Duration
		wantTimeoutEnabled bool
		wantErr            bool
	}{
		{
			name: "valid",
			opts: JobOptions{
				timeout:        0,
				timeoutEnabled: false,
			},
			timeout:            time.Second,
			wantTimeout:        time.Second,
			wantTimeoutEnabled: true,
			wantErr:            false,
		},
		{
			name: "invalid",
			opts: JobOptions{
				timeout:        0,
				timeoutEnabled: false,
			},
			timeout: 0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithJobTimeout(tt.timeout)(&tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithJobTimeout(). got err = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if got := tt.opts.timeout; got != tt.wantTimeout {
				t.Errorf("WithJobTimeout() opts.timeout = %v, want %v", got, tt.wantTimeout)
			}
			if got := tt.opts.timeoutEnabled; got != tt.wantTimeoutEnabled {
				t.Errorf("WithJobTimeout() opts.timeout = %v, want %v", got, tt.wantTimeoutEnabled)
			}
		})
	}
}

func TestWithJobTimeoutDisabled(t *testing.T) {
	tests := []struct {
		name               string
		opts               JobOptions
		wantTimeoutEnabled bool
		wantErr            bool
	}{
		{
			name: "valid",
			opts: JobOptions{
				timeoutEnabled: true,
			},
			wantTimeoutEnabled: false,
			wantErr:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithJobTimeoutDisabled()(&tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithJobTimeoutDisabled(). got err = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if got := tt.opts.timeoutEnabled; got != tt.wantTimeoutEnabled {
				t.Errorf("WithJobTimeoutDisabled() opts.timeout = %v, want %v", got, tt.wantTimeoutEnabled)
			}
		})
	}
}

func TestWithJobRequeueRetries(t *testing.T) {
	tests := []struct {
		name        string
		opts        JobOptions
		retries     int
		wantRetries int
		wantErr     bool
	}{
		{
			name: "valid",
			opts: JobOptions{
				retries: 1000,
			},
			retries:     5,
			wantRetries: 5,
			wantErr:     false,
		},
		{
			name: "invalid",
			opts: JobOptions{
				retries: 1000,
			},
			retries:     -1,
			wantErr:     true,
			wantRetries: 1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithJobRequeueRetries(tt.retries)(&tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithJobRequeueRetries(). got err = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if got := tt.opts.retries; got != tt.wantRetries {
				t.Errorf("WithJobRequeueRetries() opts.retries = %v, want %v", got, tt.wantRetries)
			}
		})
	}
}

func TestWithJobRequeuing(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name          string
		opts          JobOptions
		requeuing     bool
		wantRequeuing bool
		wantErr       bool
	}{
		{
			name: "enabled",
			opts: JobOptions{
				requeuing: false,
			},
			requeuing:     true,
			wantRequeuing: true,
			wantErr:       false,
		},
		{
			name: "disabled",
			opts: JobOptions{
				requeuing: true,
			},
			requeuing:     false,
			wantRequeuing: false,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithJobRequeuing(tt.requeuing)(&tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithJobRequeuing(). got err = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if got := tt.opts.requeuing; got != tt.wantRequeuing {
				t.Errorf("WithJobRequeuing() opts.requeuing = %v, want %v", got, tt.wantRequeuing)
			}
		})
	}
}

func TestWithJobWorkerPoolSize(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name     string
		opts     JobOptions
		size     int
		wantSize int
		wantErr  bool
	}{
		{
			name: "valid",
			opts: JobOptions{
				workerPoolSize: 10,
			},
			size:     1,
			wantSize: 1,
			wantErr:  false,
		},
		{
			name: "zero",
			opts: JobOptions{
				workerPoolSize: 10,
			},
			size:     0,
			wantSize: 10,
			wantErr:  true,
		},
		{
			name: "negative",
			opts: JobOptions{
				workerPoolSize: 10,
			},
			size:     -10,
			wantSize: 10,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WithJobWorkerPoolSize(tt.size)(&tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithJobWorkerPoolSize(). got err = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if got := tt.opts.workerPoolSize; got != tt.wantSize {
				t.Errorf("WithJobWorkerPoolSize() opts.size = %v, want %v", got, tt.wantSize)
			}
		})
	}
}
