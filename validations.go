package jobq

import (
	"errors"
	"regexp"
	"time"
)

// Validation errors
var (
	ErrJobMapUndefined        = errors.New("job map not defined")
	ErrAlreadyRegistered      = errors.New("job already registered")
	ErrInvalidRetries         = errors.New("retries should be >= 0")
	ErrInvalidPoolSize        = errors.New("pool size should be > 0")
	ErrInvalidTimeout         = errors.New("timeout should be higher than 0")
	ErrInvalidStartTime       = errors.New("start_at time should be future time")
	ErrInvalidTaskBodyScanner = errors.New("task body scanner should not be nil")
	ErrInvalidTaskBodyValuer  = errors.New("task body valuer should not be nil")
	ErrInvalidJob             = errors.New("job should not be nil")
	ErrInvalidJobName         = errors.New("invalid job name. should be snake_case")
)

const (
	jobNameRegex = "^[a-z0-9][a-z0-9_]{1,48}[a-z0-9]$"
)

func validateJobName(name string) error {
	reg := regexp.MustCompile(jobNameRegex)
	if !reg.MatchString(name) {
		return ErrInvalidJobName
	}
	return nil
}

func validateIfJobUnregistered(jobName string, jobs map[string]Job) error {
	if jobs == nil {
		return ErrJobMapUndefined
	}
	if _, ok := jobs[jobName]; ok {
		return ErrAlreadyRegistered
	}
	return nil
}

func validateJob(job Job) error {
	if job == nil {
		return ErrInvalidJob
	}
	return nil
}

func validateTaskBodyScanner(body Scanner) error {
	if body == nil {
		return ErrInvalidTaskBodyScanner
	}
	return nil
}

func validateTaskBodyValuer(body Valuer) error {
	if body == nil {
		return ErrInvalidTaskBodyValuer
	}
	return nil
}

func validateRetries(retries int) error {
	if retries < 0 {
		return ErrInvalidRetries
	}
	return nil
}

func validatePoolSize(size int) error {
	if size < 1 {
		return ErrInvalidPoolSize
	}
	return nil
}

func validateTimeout(timeout time.Duration) error {
	if timeout < 1 {
		return ErrInvalidTimeout
	}
	return nil
}

func validateStartTime(startAt time.Time) error {
	if startAt.Before(time.Now()) {
		return ErrInvalidStartTime
	}
	return nil
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
