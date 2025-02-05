# Bootstrap

A Go package for managing service lifecycle with graceful startup and shutdown capabilities.

[![Go Reference](https://pkg.go.dev/badge/github.com/n-r-w/bootstrap.svg)](https://pkg.go.dev/github.com/n-r-w/bootstrap)
![CI Status](https://github.com/n-r-w/bootstrap/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/n-r-w/bootstrap)](https://goreportcard.com/report/github.com/n-r-w/bootstrap)

## Features

- **Flexible Service Management**:
  - Ordered service startup and shutdown
  - Parallel service startup for unordered services
  - After-start services that run after main services
  - Graceful shutdown handling

- **Health Checks**:
  - Readiness checks for service availability
  - Liveness checks for service health
  - Custom error handling for failed health checks

- **Resilient Operations**:
  - Configurable restart policies using backoff strategies
  - Graceful shutdown with configurable timeouts
  - Context-based cancellation support

- **Extensible Logging**:  
  - Context-aware logging methods
  - Flexible logger implementation

## Installation

```bash
go get github.com/n-r-w/bootstrap
```

## Usage

See [example](example/main.go)

## Configuration Options

- `WithStartTimeout(timeout time.Duration)`: Sets the timeout for service startup
- `WithStopTimeout(timeout time.Duration)`: Sets the timeout for graceful shutdown
- `WithHealthCheck(checker IHealthChecker)`: Configures health checking
- `WithOrdered(services ...IService)`: Adds services that must start in order
- `WithUnordered(services ...IService)`: Adds services that can start in parallel
- `WithAfterStart(services ...IService)`: Adds services to start after all others
- `WithRunFunc(func(context.Context) error)`: Sets a function to run after services start
- `WithLogger(logger ILogger)`: Configures a custom logger

## Service Interface

Services must implement the `IService` interface:

```go
type IService interface {
    Info() Info
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

The `Info` struct contains:

- `Name`: Service identifier
- `RestartPolicy`: Optional backoff configuration for restart attempts

## Health Checks

Health checks can be added using the `IHealthChecker` interface

## Executor package

[Executor](executor/service.go) package provides an interface for executing a function at specified time intervals.
Supports bootstrap.IService interface.
