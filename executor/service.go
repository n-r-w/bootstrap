package executor

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/n-r-w/bootstrap"
)

// OnErrorFunc - function for handling errors.
type OnErrorFunc func(ctx context.Context, err error)

// Option function for configuring a Service
type Option func(*Service)

// WithOnError - sets function for handling errors
func WithOnError(onErrorFunc OnErrorFunc) Option {
	return func(s *Service) {
		s.onErrorFunc = onErrorFunc
	}
}

// WithJitter - sets jitter
func WithJitter(jitter time.Duration) Option {
	return func(s *Service) {
		s.jitter = jitter
	}
}

// Service - execute background process with interval.
type Service struct {
	name string

	e        IExecutor
	interval time.Duration
	jitter   time.Duration
	wg       sync.WaitGroup

	onErrorFunc OnErrorFunc

	cancelFunc context.CancelFunc
}

var _ bootstrap.IService = (*Service)(nil)

// New creates a new service instance.
func New(name string, e IExecutor, interval time.Duration, opts ...Option) (*Service, error) {
	if name == "" {
		return nil, errors.New("name is empty")
	}
	if e == nil {
		return nil, errors.New("executor is nil")
	}
	if interval <= 0 {
		return nil, errors.New("interval must be positive")
	}

	s := &Service{
		name:     name,
		e:        e,
		interval: interval,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

// Info returns service information.
func (s *Service) Info() bootstrap.Info {
	return bootstrap.Info{
		Name: s.name,
	}
}

// Start runs the service in the background.
func (s *Service) Start(ctx context.Context) error {
	// ignore incoming context timeout, since the method Start receives a context
	// with a short lifetime (time to start the service)
	ctx, s.cancelFunc = context.WithCancel(context.WithoutCancel(ctx))

	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := s.e.Execute(ctx); err != nil {
					if s.onErrorFunc != nil {
						s.onErrorFunc(ctx, err)
					}
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(s.interval + time.Duration(rand.Float64()*float64(s.jitter))): //nolint:gosec // ok
				}
			}
		}
	}()

	return nil
}

// Stop stops the service.
func (s *Service) Stop(ctx context.Context) error {
	err := s.e.StopExecutor(ctx)
	s.cancelFunc()
	s.wg.Wait()
	return err
}
