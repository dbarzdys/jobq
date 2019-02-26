package jobq

import (
	"errors"
	"testing"
	"time"
)

func Test_validateJobName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		job  string
		want error
	}{
		{
			name: "valid",
			job:  "testjob_v1",
			want: nil,
		},
		{
			name: "too_short",
			job:  "t",
			want: ErrInvalidJobName,
		},
		{
			name: "too_long",
			job:  "ttttttttttttttttttttttttttttttttttttttttttttttttttt",
			want: ErrInvalidJobName,
		},
		{
			name: "invalid_snake_case_1",
			job:  "_testjob",
			want: ErrInvalidJobName,
		},
		{
			name: "invalid_snake_case_2",
			job:  "testjob_",
			want: ErrInvalidJobName,
		},
		{
			name: "invalid_snake_case_3",
			job:  "Test_Job",
			want: ErrInvalidJobName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateJobName(tt.job); err != tt.want {
				t.Errorf("validateJobName() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_validateIfJobUnregistered(t *testing.T) {
	tests := []struct {
		name    string
		jobName string
		jobs    map[string]Job
		want    error
	}{
		{
			name:    "map_undefined",
			jobName: "test",
			jobs:    nil,
			want:    ErrJobMapUndefined,
		},
		{
			name:    "already_registered",
			jobName: "test",
			jobs:    map[string]Job{"test": nil},
			want:    ErrAlreadyRegistered,
		},
		{
			name:    "valid",
			jobName: "test",
			jobs:    map[string]Job{},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateIfJobUnregistered(tt.jobName, tt.jobs); err != tt.want {
				t.Errorf("validateIfJobUnregistered() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_validateJob(t *testing.T) {
	tests := []struct {
		name string
		job  Job
		want error
	}{
		{
			name: "invalid",
			job:  nil,
			want: ErrInvalidJob,
		},
		{
			name: "valid",
			job:  mockJob{},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateJob(tt.job); err != tt.want {
				t.Errorf("validateJob() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_validateTaskBodyScanner(t *testing.T) {
	tests := []struct {
		name    string
		scanner Scanner
		want    error
	}{
		{
			name:    "invalid",
			scanner: nil,
			want:    ErrInvalidTaskBodyScanner,
		},
		{
			name:    "valid",
			scanner: mockScanner{},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTaskBodyScanner(tt.scanner); err != tt.want {
				t.Errorf("validateTaskBodyScanner() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_validateTaskBodyValuer(t *testing.T) {
	tests := []struct {
		name   string
		valuer Valuer
		want   error
	}{
		{
			name:   "invalid",
			valuer: nil,
			want:   ErrInvalidTaskBodyValuer,
		},
		{
			name:   "valid",
			valuer: mockValuer{},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTaskBodyValuer(tt.valuer); err != tt.want {
				t.Errorf("validateTaskBodyValuer() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_validateRetries(t *testing.T) {
	tests := []struct {
		name    string
		retries int
		want    error
	}{
		{
			name:    "negative",
			retries: -1,
			want:    ErrInvalidRetries,
		},
		{
			name:    "positive",
			retries: 1,
			want:    nil,
		},
		{
			name:    "zero",
			retries: 0,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateRetries(tt.retries); err != tt.want {
				t.Errorf("validateRetries() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_validatePoolSize(t *testing.T) {
	tests := []struct {
		name string
		size int
		want error
	}{
		{
			name: "negative",
			size: -1,
			want: ErrInvalidPoolSize,
		},
		{
			name: "zero",
			size: 0,
			want: ErrInvalidPoolSize,
		},
		{
			name: "positive",
			size: 1,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePoolSize(tt.size); err != tt.want {
				t.Errorf("validatePoolSize() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_validateTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		want    error
	}{
		{
			name:    "negative",
			timeout: -time.Second,
			want:    ErrInvalidTimeout,
		},
		{
			name:    "zero",
			timeout: 0,
			want:    ErrInvalidTimeout,
		},
		{
			name:    "positive",
			timeout: time.Second,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTimeout(tt.timeout); err != tt.want {
				t.Errorf("validateTimeout() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_validateStartTime(t *testing.T) {
	tests := []struct {
		name    string
		startAt time.Time
		want    error
	}{
		{
			name:    "past",
			startAt: time.Now().Add(-time.Minute),
			want:    ErrInvalidStartTime,
		},
		{
			name:    "future",
			startAt: time.Now().Add(time.Minute),
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateStartTime(tt.startAt); err != tt.want {
				t.Errorf("validateStartTime() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func Test_firstError(t *testing.T) {
	var (
		errOne   = errors.New("one")
		errTwo   = errors.New("two")
		errThree = errors.New("three")
	)
	tests := []struct {
		name string
		errs []error
		want error
	}{
		{
			name: "empty",
			errs: nil,
			want: nil,
		},
		{
			name: "no_errs",
			errs: []error{nil, nil, nil},
			want: nil,
		},
		{
			name: "first",
			errs: []error{errOne, errTwo, errThree},
			want: errOne,
		},
		{
			name: "second",
			errs: []error{nil, errTwo, errThree},
			want: errTwo,
		},
		{
			name: "third",
			errs: []error{nil, nil, errThree},
			want: errThree,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := firstError(tt.errs...); err != tt.want {
				t.Errorf("firstError() error = %v, want %v", err, tt.want)
			}
		})
	}
}
