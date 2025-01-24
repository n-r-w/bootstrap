//nolint:mnd // example code
package main

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/n-r-w/bootstrap"
)

// ExampleService - demonstrates a basic service implementation.
type ExampleService struct {
	name    string
	running chan struct{}
}

// NewExampleService - creates new example service.
func NewExampleService(name string) *ExampleService {
	return &ExampleService{
		name:    name,
		running: make(chan struct{}),
	}
}

// Info - returns service information.
func (s *ExampleService) Info() bootstrap.Info {
	// Service with restart policy that will retry with exponential backoff
	return bootstrap.Info{
		Name: s.name,
		RestartPolicy: []backoff.RetryOption{
			backoff.WithBackOff(backoff.NewExponentialBackOff()),
			backoff.WithMaxElapsedTime(5 * time.Second),
		},
	}
}

// Start - starts the service.
func (s *ExampleService) Start(ctx context.Context) error {
	// Simulate some startup work
	time.Sleep(time.Second)

	close(s.running)
	return nil
}

// Stop - stops the service.
func (s *ExampleService) Stop(ctx context.Context) error {
	return nil
}
