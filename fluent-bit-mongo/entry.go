package main

import (
	"errors"
	"fmt"
)

type PodLog struct {
	Log       string
	Label     map[string]string
	Namespace string
	PodID     string
	PodName   string
	Time      uint64
}

var ErrNoRecord = errors.New("failed to decode entry")

type ErrRetry struct {
	Cause error
}

func (err *ErrRetry) Error() string {
	return fmt.Sprintf("retry: %s", err.Cause)
}

func (err *ErrRetry) Unwrap() error {
	return err.Cause
}

func (err *ErrRetry) Is(err2 error) bool {
	_, ok := err2.(*ErrRetry)

	return ok
}
