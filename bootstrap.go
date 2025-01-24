package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cenkalti/backoff/v5"
	"golang.org/x/sync/errgroup"
)

// Option - function for configuring Bootstrap.
type Option func(*Bootstrap)

// WithStopTimeout - sets the timeout for graceful shutdown.
func WithStopTimeout(timeout time.Duration) Option {
	return func(b *Bootstrap) {
		b.stopTimeout = timeout
	}
}

// WithStartTimeout - sets the timeout for graceful startup.
func WithStartTimeout(timeout time.Duration) Option {
	return func(b *Bootstrap) {
		b.startTimeout = timeout
	}
}

// WithHealthCheck - sets health.Handler for service readiness check.
func WithHealthCheck(hch IHealthChecher) Option {
	return func(b *Bootstrap) {
		b.hch = hch
	}
}

// WithOrdered - sets services that must be started in a specific order.
// Shutdown happens in reverse order after stopping services that don't require ordering.
func WithOrdered(services ...IService) Option {
	return func(b *Bootstrap) {
		b.ordered = services
	}
}

// WithUnordered - sets services that can be started in parallel.
// Shutdown also happens in parallel before stopping services that require ordering.
func WithUnordered(services ...IService) Option {
	return func(b *Bootstrap) {
		b.unordered = services
	}
}

// WithAfterStart - sets services that should be started after all others.
// Startup happens in order after starting all other services.
// Shutdown happens in reverse order before stopping all other services.
func WithAfterStart(services ...IService) Option {
	return func(b *Bootstrap) {
		b.afterStart = services
	}
}

// WithRunFunc - sets a function to run. If set, the function will be called and then work will be completed.
func WithRunFunc(runFunc func(context.Context) error) Option {
	return func(b *Bootstrap) {
		b.runFunc = runFunc
	}
}

// WithLogger - sets logger for logging.
func WithLogger(logger ILogger) Option {
	return func(b *Bootstrap) {
		b.logger = logger
	}
}

// Bootstrap - helper for starting services.
type Bootstrap struct {
	appName    string
	unordered  []IService
	ordered    []IService
	afterStart []IService

	stopTimeout  time.Duration
	startTimeout time.Duration

	hch    IHealthChecher
	logger ILogger

	runFunc func(context.Context) error
}

// New - creates a new Bootstrap instance.
func New(appName string) (*Bootstrap, error) {
	if appName == "" {
		return nil, errors.New("app name is empty")
	}

	return &Bootstrap{
		appName: appName,
	}, nil
}

// Run - starts all services and waits for stop signal.
func (b *Bootstrap) Run(ctx context.Context, opts ...Option) error {
	for _, opt := range opts {
		opt(b)
	}

	b.logger = &loggerWrapper{logger: b.logger}

	var ready int32
	if b.hch != nil {
		b.hch.AddLivenessCheck(fmt.Sprintf("%s-app", b.appName), func() error {
			b.logger.Debugf(ctx, "checking liveness")

			if atomic.LoadInt32(&ready) == 1 {
				return nil
			}
			return fmt.Errorf("app %s is not ready yet", b.appName)
		})

		b.hch.AddCheckErrorHandler(func(name string, err error) {
			b.logger.Errorf(ctx, "healthcheck failed for %s: %v", name, err)
		})
	}

	if err := b.startHelper(ctx); err != nil {
		return err
	}

	atomic.StoreInt32(&ready, 1)

	if b.runFunc != nil {
		b.logger.Infof(ctx, "worker func started")
		started := time.Now()

		if err := b.runFunc(ctx); err != nil {
			b.logger.Errorf(ctx, "worker func failed. duration: %v, error: %v", time.Since(started), err)
		} else {
			b.logger.Infof(ctx, "worker func finished. duration: %v", time.Since(started))
		}
	} else {
		b.waitInterrupt(ctx)
	}

	b.stopHelper(ctx)

	return nil
}

// startHelper - starts all services.
func (b *Bootstrap) startHelper(ctx context.Context) error {
	var startCtx context.Context

	if b.startTimeout > 0 {
		var startCancelFunc context.CancelFunc
		startCtx, startCancelFunc = context.WithTimeout(ctx, b.startTimeout)
		defer startCancelFunc()
	} else {
		startCtx = ctx
	}

	started := make([]IService, 0, len(b.ordered)+len(b.unordered)+len(b.afterStart))
	checkStarted := func(s IService) error {
		for _, startedService := range started {
			if startedService.Info().Name == s.Info().Name {
				return fmt.Errorf("service %s is already started", s.Info().Name)
			}
		}
		started = append(started, s)
		return nil
	}

	// first start services that require ordering
	for _, s := range b.ordered {
		if err := checkStarted(s); err != nil {
			return err
		}

		if err := b.startService(startCtx, s); err != nil {
			return err
		}
	}

	// then start remaining services in parallel
	egStart := errgroup.Group{}
	for _, s := range b.unordered {
		if err := checkStarted(s); err != nil {
			return err
		}

		s1 := s
		egStart.Go(func() error {
			return b.startService(startCtx, s1)
		})
	}

	if err := egStart.Wait(); err != nil {
		return err
	}

	b.logger.Infof(ctx, "main services started")

	// then start services that should be started after all others in order
	for _, s := range b.afterStart {
		if err := checkStarted(s); err != nil {
			return err
		}

		if err := b.startService(startCtx, s); err != nil {
			return err
		}
	}

	b.logger.Infof(ctx, "all services started")

	return nil
}

// startService - starts a service.
func (b *Bootstrap) startService(ctx context.Context, s IService) error {
	b.logger.Infof(ctx, "starting service. name: %s", s.Info().Name)

	var started int32
	if b.hch != nil {
		b.hch.AddReadinessCheck(s.Info().Name, func() error {
			b.logger.Debugf(ctx, "checking readiness. name: %s", s.Info().Name)

			if atomic.LoadInt32(&started) == 0 {
				return fmt.Errorf("service %s is not started yet", s.Info().Name)
			}
			return nil
		})
	}

	var err error
	if s.Info().RestartPolicy == nil {
		err = s.Start(ctx)
	} else {
		operation := func() (struct{}, error) {
			if ctx.Err() != nil {
				return struct{}{}, backoff.Permanent(ctx.Err())
			}

			return struct{}{}, s.Start(ctx)
		}

		_, err = backoff.Retry(ctx, operation,
			backoff.WithBackOff(s.Info().RestartPolicy),
			backoff.WithNotify(func(err error, d time.Duration) {
				b.logger.Warningf(ctx, "failed to run service. retrying. name: %s, error: %v, delay: %v", s.Info().Name, err, d)
			}),
		)
	}

	if err != nil {
		return fmt.Errorf("failed to run service %s: %w", s.Info().Name, err)
	}
	b.logger.Infof(ctx, "service started. name: %s", s.Info().Name)

	atomic.StoreInt32(&started, 1)
	return nil
}

// waitInterrupt - waits for OS termination signal or context cancellation.
func (b *Bootstrap) waitInterrupt(ctx context.Context) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		b.logger.Infof(ctx, "got OS terminate signal, stopping")
	case <-ctx.Done():
		b.logger.Infof(ctx, "bootstrap context done, stopping")
	}
}

// stopHelper - performs graceful shutdown of all services.
func (b *Bootstrap) stopHelper(ctx context.Context) {
	var stopCtx context.Context
	if b.stopTimeout > 0 {
		var stopCancelFunc context.CancelFunc
		stopCtx, stopCancelFunc = context.WithTimeout(ctx, b.stopTimeout)
		defer stopCancelFunc()
	} else {
		stopCtx = ctx
	}

	stopFunc := func(s IService) {
		b.logger.Infof(ctx, "stopping service. name: %s", s.Info().Name)
		if err := s.Stop(stopCtx); err != nil {
			b.logger.Errorf(ctx, "failed to stop service. name: %s, error: %v", s.Info().Name, err)
		} else {
			b.logger.Infof(ctx, "service stopped. name: %s", s.Info().Name)
		}
	}

	// first stop afterStart services since they were started after all others
	for i := len(b.afterStart) - 1; i >= 0; i-- {
		stopFunc(b.afterStart[i])
	}

	// stop services that don't require ordering since they were started after those that require ordering
	var wgStop sync.WaitGroup
	wgStop.Add(len(b.unordered))
	for _, s := range b.unordered {
		go func(srv IService) {
			defer wgStop.Done()
			stopFunc(srv)
		}(s)
	}
	wgStop.Wait()

	// then stop services that require ordering in reverse start order
	for i := len(b.ordered) - 1; i >= 0; i-- {
		stopFunc(b.ordered[i])
	}

	b.logger.Infof(ctx, "all services stopped")
}
