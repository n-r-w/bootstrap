//nolint:mnd // example code
package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/n-r-w/bootstrap"
)

func main() {
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

	// set start t
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Run bootstrap with options
		err = bs.Run(ctx,
			// Set custom logger
			bootstrap.WithLogger(newCustomLogger()),
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
					log.Println("All services started, running main logic...")

					// Simulate some work
					time.Sleep(2 * time.Second)

					log.Println("Main logic completed")
					return nil
				}),
			*/
		)
		if err != nil {
			log.Printf("Bootstrap failed: %v\n", err)
		}
	}()

	// Wait for context cancellation
	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Println("Waiting for context cancellation...")
		time.Sleep(5 * time.Second)

		cancel()
		log.Println("Context cancelled")
	}()

	// Wait for bootstrap to finish
	wg.Wait()
}
