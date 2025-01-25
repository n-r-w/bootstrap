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
	"github.com/n-r-w/ctxlog"
	"golang.org/x/sync/errgroup"
)

// Bootstrap - helper for starting services.
type Bootstrap struct {
	appName    string
	unordered  []IService
	ordered    []IService
	afterStart []IService

	stopTimeout  time.Duration
	startTimeout time.Duration

	hch    IHealthChecher
	logger ctxlog.ILogger

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

	if b.logger == nil {
		b.logger = ctxlog.NewStubWrapper()
	}

	var ready int32
	if b.hch != nil {
		b.hch.AddLivenessCheck(fmt.Sprintf("%s-app", b.appName), func() error {
			b.logger.Debug(ctx, "checking liveness status")

			if atomic.LoadInt32(&ready) == 1 {
				return nil
			}
			return fmt.Errorf("app %s is not ready yet", b.appName)
		})

		b.hch.AddCheckErrorHandler(func(name string, err error) {
			b.logger.Error(ctx, "healthcheck failed", "name", name, "error", err)
		})
	}

	if err := b.startHelper(ctx); err != nil {
		return err
	}

	atomic.StoreInt32(&ready, 1)

	if b.runFunc != nil {
		b.logger.Info(ctx, "worker function started")
		started := time.Now()

		if err := b.runFunc(ctx); err != nil {
			b.logger.Error(ctx, "worker function execution failed", "duration", time.Since(started), "error", err)
		} else {
			b.logger.Info(ctx, "worker function completed", "duration", time.Since(started))
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

	b.logger.Info(ctx, "main services started")

	// then start services that should be started after all others in order
	for _, s := range b.afterStart {
		if err := checkStarted(s); err != nil {
			return err
		}

		if err := b.startService(startCtx, s); err != nil {
			return err
		}
	}

	b.logger.Info(ctx, "all services started")

	return nil
}

// startService - starts a service.
func (b *Bootstrap) startService(ctx context.Context, s IService) error {
	b.logger.Info(ctx, "starting service", "name", s.Info().Name)

	var started int32
	if b.hch != nil {
		b.hch.AddReadinessCheck(s.Info().Name, func() error {
			b.logger.Debug(ctx, "checking service readiness", "name", s.Info().Name)

			if atomic.LoadInt32(&started) == 0 {
				return fmt.Errorf("service %s is not started yet", s.Info().Name)
			}
			return nil
		})
	}

	var err error
	if len(s.Info().RestartPolicy) == 0 {
		err = s.Start(ctx)
	} else {
		operation := func() (struct{}, error) {
			if ctx.Err() != nil {
				return struct{}{}, backoff.Permanent(ctx.Err())
			}

			return struct{}{}, s.Start(ctx)
		}

		retryOptions := s.Info().RestartPolicy
		retryOptions = append(retryOptions, backoff.WithNotify(func(err error, d time.Duration) {
			b.logger.Warn(ctx, "failed to run service - retrying", "name", s.Info().Name, "error", err, "delay", d)
		}))

		_, err = backoff.Retry(ctx, operation, retryOptions...)
	}

	if err != nil {
		return fmt.Errorf("failed to run service %s: %w", s.Info().Name, err)
	}
	b.logger.Info(ctx, "service started", "name", s.Info().Name)

	atomic.StoreInt32(&started, 1)
	return nil
}

// waitInterrupt - waits for OS termination signal or context cancellation.
func (b *Bootstrap) waitInterrupt(ctx context.Context) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigs:
		b.logger.Info(ctx, "received OS terminate signal, shutting down")
	case <-ctx.Done():
		b.logger.Info(ctx, "bootstrap context done, shutting down")
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
		b.logger.Info(ctx, "stopping service", "name", s.Info().Name)
		if err := s.Stop(stopCtx); err != nil {
			b.logger.Error(ctx, "failed to stop service", "name", s.Info().Name, "error", err)
		} else {
			b.logger.Info(ctx, "service stopped", "name", s.Info().Name)
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

	b.logger.Info(ctx, "all services stopped")
}
