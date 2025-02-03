package bootstrap

import (
	"context"
	"time"

	"github.com/n-r-w/ctxlog"
)

// CleanUpFunc - function for cleaning up resources.
type CleanUpFunc func() error

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
func WithLogger(logger ctxlog.ILogger) Option {
	return func(b *Bootstrap) {
		b.logger = logger
	}
}

// WithCleanUp - adds a function to be called during shutdown.
// Called in reverse order.
func WithCleanUp(cleanUp ...CleanUpFunc) Option {
	return func(b *Bootstrap) {
		b.cleanUp = cleanUp
	}
}
