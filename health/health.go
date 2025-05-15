package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/n-r-w/bootstrap"
)

// http://host:port/readiness?full=1 for full output

const (
	// LivenessHandlerPath - liveness handler path.
	LivenessHandlerPath = "/liveness"
	// ReadinessHandlerPath - readiness handler path.
	ReadinessHandlerPath = "/readiness"

	successCheckerResultString = "OK"
)

// New creates new Health, which implements http.Handler + IHealther and allows
// adding liveness and readiness checks.
func New() *Health {
	h := &Health{
		livenessChecks:  make(map[string]bootstrap.CheckFunc),
		readinessChecks: make(map[string]bootstrap.CheckFunc),
	}
	h.Handle(LivenessHandlerPath, http.HandlerFunc(h.LiveEndpoint))
	h.Handle(ReadinessHandlerPath, http.HandlerFunc(h.ReadyEndpoint))

	return h
}

// Health implements http.Health and allows adding liveness and readiness checks.
type Health struct {
	http.ServeMux
	checksMutex     sync.RWMutex
	livenessChecks  map[string]bootstrap.CheckFunc
	readinessChecks map[string]bootstrap.CheckFunc
	errorHandler    bootstrap.ErrorHandlerFunc
}

var _ bootstrap.IHealthChecher = (*Health)(nil)

// LiveEndpoint is an HTTP handler for the /live endpoint, which
// is useful if you need to add it to your own HTTP handler tree.
func (s *Health) LiveEndpoint(w http.ResponseWriter, r *http.Request) {
	s.handle(w, r, s.livenessChecks)
}

// ReadyEndpoint is an HTTP handler for the /ready endpoint, which
// is useful if you need to add it to your own HTTP handler tree.
func (s *Health) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {
	s.handle(w, r, s.readinessChecks, s.livenessChecks)
}

// AddLivenessCheck adds a check indicating that this instance
// of the application should be destroyed or restarted. A failed liveness
// check indicates that this instance is not working.
// Each liveness check is also included as a readiness check.
func (s *Health) AddLivenessCheck(name string, check bootstrap.CheckFunc) {
	s.checksMutex.Lock()
	defer s.checksMutex.Unlock()
	s.livenessChecks[name] = check
}

// AddReadinessCheck adds a check indicating that this instance
// of the application currently cannot serve requests due to an external
// dependency or some temporary failure. If the readiness check fails, this instance
// should no longer receive requests, but it should not be restarted or
// destroyed.
func (s *Health) AddReadinessCheck(name string, check bootstrap.CheckFunc) {
	s.checksMutex.Lock()
	defer s.checksMutex.Unlock()
	s.readinessChecks[name] = check
}

// AddCheckErrorHandler adds a callback for handling a failed
// check (for the purpose of logging errors, etc).
// The function should not block the execution thread for a long time.
func (s *Health) AddCheckErrorHandler(handler bootstrap.ErrorHandlerFunc) {
	s.errorHandler = handler
}

type result struct {
	name   string
	result string
}

func (s *Health) collectChecks(checks map[string]bootstrap.CheckFunc, resultsOut map[string]string) (status int) {
	s.checksMutex.RLock()
	defer s.checksMutex.RUnlock()

	status = http.StatusOK

	limit := len(checks)

	if limit == 0 {
		return status
	}

	results := make(chan result, limit)

	for name, check := range checks {
		go func(name string, check bootstrap.CheckFunc) {
			defer func() {
				if r := recover(); r != nil {
					results <- result{
						name:   name,
						result: fmt.Sprintf("checker panic recovered: %v", r),
					}

					if s.errorHandler != nil {
						s.errorHandler(name, fmt.Errorf("checker panic recovered: %v", r))
					}
				}
			}()

			val := successCheckerResultString
			if err := check(); err != nil {
				val = err.Error()

				if s.errorHandler != nil {
					s.errorHandler(name, err)
				}
			}

			results <- result{
				name:   name,
				result: val,
			}
		}(name, check)
	}

	for res := range results {
		resultsOut[res.name] = res.result

		if res.result != successCheckerResultString {
			status = http.StatusServiceUnavailable
		}

		limit--

		if limit == 0 {
			close(results)
		}
	}

	return status
}

func (s *Health) handle(w http.ResponseWriter, r *http.Request, checks ...map[string]bootstrap.CheckFunc) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	checkResults := make(map[string]string)
	status := http.StatusOK
	for _, m := range checks {
		if st := s.collectChecks(m, checkResults); st != http.StatusOK {
			status = st
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(status)

	// if no ?full=1, return empty body. Kubernetes cares only about
	// HTTP status codes, so we won't waste bytes on the full request body.
	if r.URL.Query().Get("full") != "1" {
		_, _ = w.Write([]byte("{}\n"))
		return
	}

	// Write the JSON body, ignoring any encoding errors (which are
	// actually impossible because we're encoding a map[string]string).
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	_ = encoder.Encode(checkResults)
}
