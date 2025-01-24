package bootstrap

//go:generate mockgen -source interface.go -destination interface_mock.go -package bootstrap

import (
	"context"

	"github.com/cenkalti/backoff/v5"
)

// Info - service information.
type Info struct {
	// Name - service name.
	Name string
	// RestartPolicy - service restart policy on error. If empty, the service will not be restarted.
	RestartPolicy []backoff.RetryOption
}

// IService - interface for a service that can be started and stopped.
type IService interface {
	// Info - returns service information.
	Info() Info
	// Start - starts the service.
	// WARNING:
	// The passed context contains a timeout for starting the service and will be canceled.
	// Therefore, it should not be used to create objects with a long lifecycle.
	Start(ctx context.Context) error
	// Stop - stops the service.
	Stop(ctx context.Context) error
}

// IHealthChecher interface for health checks.
type IHealthChecher interface {
	// AddReadinessCheck adds a check indicating that this instance
	// of the application currently cannot serve requests due to an external
	// dependency or some temporary failure. If the readiness check fails, this instance
	// should no longer receive requests, but it should not be restarted or
	// destroyed.
	AddReadinessCheck(name string, check func() error)
	// AddLivenessCheck adds a check indicating that this instance
	// of the application should be destroyed or restarted. A failed liveness
	// check indicates that this instance is not working.
	// Each liveness check is also included as a readiness check.
	AddLivenessCheck(name string, check func() error)
	// AddCheckErrorHandler adds a callback for handling a failed
	// check (for the purpose of logging errors, etc).
	// The function should not block the execution thread for a long time.
	AddCheckErrorHandler(handler func(name string, err error))
}

// ILogger interface for logging.
type ILogger interface {
	Debugf(ctx context.Context, format string, args ...any)
	Infof(ctx context.Context, format string, args ...any)
	Warningf(ctx context.Context, format string, args ...any)
	Errorf(ctx context.Context, format string, args ...any)
}
