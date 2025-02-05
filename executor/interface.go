package executor

//go:generate mockgen -source interface.go -destination interface_mock.go -package executor

import "context"

// IExecutor interface for executing an operation at specified intervals
type IExecutor interface {
	// Execute runs the operation
	Execute(ctx context.Context) error
	// StopExecutor stops the executor. For correct service shutdown.
	StopExecutor(ctx context.Context) error
}
