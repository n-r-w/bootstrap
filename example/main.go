//nolint:mnd // example code
package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/n-r-w/bootstrap"
	"github.com/n-r-w/ctxlog"
)

func main() {
	// Create custom logger and put it into context
	ctx := ctxlog.MustContext(
		context.Background(),
		ctxlog.WithEnvType(ctxlog.EnvProduction),
	)

	// Create bootstrap instance
	bs, err := bootstrap.New("example-app")
	if err != nil {
		log.Fatalf("Failed to create bootstrap: %v", err)
	}

	// Create services
	db := NewExampleService("database")
	cache := NewExampleService("cache")
	api := NewExampleService("api")
	worker := NewExampleService("background-worker")

	// set start timeout
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Run bootstrap with options
		err = bs.Run(ctx,
			// Set custom logger (ctxlog.NewWrapper is a wrapper around slog+zap, that implements ctxlog.ILogger)
			bootstrap.WithLogger(ctxlog.NewWrapper()),
			// Set timeouts
			bootstrap.WithStartTimeout(5*time.Second),
			bootstrap.WithStopTimeout(5*time.Second),

			// Database and cache must start in order
			bootstrap.WithOrdered(db, cache),

			// API and worker can start in parallel after db and cache
			bootstrap.WithUnordered(api, worker),

			// Optional: add services that should start after everything else
			// bootstrap.WithAfterStart(metrics),

			// Optional: add a function to run after all services are started
			/*
				bootstrap.WithRunFunc(func(ctx context.Context) error {
					ctxlog.Info(ctx, "All services started, running main logic...")

					// Simulate some work
					time.Sleep(2 * time.Second)

					ctxlog.Info(ctx, "Main logic completed")
					return nil
				}),
			*/
		)
		if err != nil {
			ctxlog.Info(ctx, "Bootstrap failed", "error", err)
		}
	}()

	// Wait for context cancellation
	wg.Add(1)
	go func() {
		defer wg.Done()

		ctxlog.Info(ctx, "Waiting for context cancellation...")
		time.Sleep(5 * time.Second)

		cancel()
		ctxlog.Info(ctx, "Context cancelled")
	}()

	// Wait for bootstrap to finish
	wg.Wait()
}
