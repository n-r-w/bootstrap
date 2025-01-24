package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/stretchr/testify/require"
)

type testServiceNoBackoff struct {
	name string
	data *testData
}

func (s *testServiceNoBackoff) Info() Info {
	return Info{
		Name: s.name,
	}
}

func (s *testServiceNoBackoff) Start(_ context.Context) error {
	s.data.muNoBackoff.Lock()
	defer s.data.muNoBackoff.Unlock()
	s.data.startNoBackoff = append(s.data.startNoBackoff, s.name)
	return nil
}

func (s *testServiceNoBackoff) Stop(_ context.Context) error {
	s.data.muNoBackoff.Lock()
	defer s.data.muNoBackoff.Unlock()
	s.data.stopNoBackoff = append(s.data.stopNoBackoff, s.name)

	return nil
}

type testData struct {
	startNoBackoff []string
	stopNoBackoff  []string
	muNoBackoff    sync.Mutex
}

// TestRunNoBackoff - verifies that services without backoff are started.
func TestRunNoBackoff(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	b, err := New("test")
	require.NoError(t, err)

	data := &testData{}

	var (
		errRun error
		wg     sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		defer wg.Done()

		errRun = b.Run(ctx,
			WithUnordered(&testServiceNoBackoff{"uo1", data}, &testServiceNoBackoff{"uo2", data}),
			WithOrdered(&testServiceNoBackoff{"o1", data}, &testServiceNoBackoff{"o2", data}),
			WithAfterStart(&testServiceNoBackoff{"as1", data}, &testServiceNoBackoff{"as2", data}),
		)
	}()

	cancel()
	wg.Wait()

	require.NoError(t, errRun)

	// startup order
	require.Len(t, data.startNoBackoff, 6)

	require.Equal(t, "o1", data.startNoBackoff[0])
	require.Equal(t, "o2", data.startNoBackoff[1])

	require.Equal(t, "as1", data.startNoBackoff[4])
	require.Equal(t, "as2", data.startNoBackoff[5])

	require.Contains(t, data.startNoBackoff, "uo1")
	require.Contains(t, data.startNoBackoff, "uo2")

	// shutdown order
	require.Len(t, data.stopNoBackoff, 6)

	require.Equal(t, "o1", data.stopNoBackoff[5])
	require.Equal(t, "o2", data.stopNoBackoff[4])

	require.Equal(t, "as1", data.stopNoBackoff[1])
	require.Equal(t, "as2", data.stopNoBackoff[0])

	require.Contains(t, data.stopNoBackoff, "uo1")
	require.Contains(t, data.stopNoBackoff, "uo2")
}

// TestRunFunc - verifies service startup using a function.
func TestRunFunc(t *testing.T) {
	ctx := context.Background()

	b, err := New("test")
	require.NoError(t, err)

	data := &testData{}

	err = b.Run(ctx,
		WithUnordered(&testServiceNoBackoff{"uo1", data}, &testServiceNoBackoff{"uo2", data}),
		WithOrdered(&testServiceNoBackoff{"o1", data}, &testServiceNoBackoff{"o2", data}),
		WithAfterStart(&testServiceNoBackoff{"as1", data}, &testServiceNoBackoff{"as2", data}),
		WithRunFunc(func(_ context.Context) error { return errors.New("test error") }),
	)

	require.NoError(t, err)

	// startup order
	require.Len(t, data.startNoBackoff, 6)

	require.Equal(t, "o1", data.startNoBackoff[0])
	require.Equal(t, "o2", data.startNoBackoff[1])

	require.Equal(t, "as1", data.startNoBackoff[4])
	require.Equal(t, "as2", data.startNoBackoff[5])

	require.Contains(t, data.startNoBackoff, "uo1")
	require.Contains(t, data.startNoBackoff, "uo2")

	// shutdown order
	require.Len(t, data.stopNoBackoff, 6)

	require.Equal(t, "o1", data.stopNoBackoff[5])
	require.Equal(t, "o2", data.stopNoBackoff[4])

	require.Equal(t, "as1", data.stopNoBackoff[1])
	require.Equal(t, "as2", data.stopNoBackoff[0])

	require.Contains(t, data.stopNoBackoff, "uo1")
	require.Contains(t, data.stopNoBackoff, "uo2")
}

// TestRunNOK - verifies that services with identical names are not started.
func TestRunNOK(t *testing.T) {
	ctx := context.Background()

	_, err := New("")
	require.Error(t, err)

	b, err := New("test")
	require.NoError(t, err)

	data := &testData{}

	// duplicate name
	require.Error(t, b.Run(ctx,
		WithUnordered(&testServiceNoBackoff{"s1", data}),
		WithOrdered(&testServiceNoBackoff{"s1", data})),
	)
}

var readyBackoff int32

type testServiceBackoff struct {
	name string
}

func (s *testServiceBackoff) Info() Info {
	return Info{
		Name:          s.name,
		RestartPolicy: backoff.NewConstantBackOff(100 * time.Millisecond),
	}
}

func (s *testServiceBackoff) Start(_ context.Context) error {
	counter := atomic.AddInt32(&readyBackoff, 1)
	if counter < 3 {
		return fmt.Errorf("restart %d", counter)
	}

	return nil
}

func (s *testServiceBackoff) Stop(_ context.Context) error {
	return nil
}

// TestRunBackoff - verifies that services with backoff are restarted.
func TestRunBackoff(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	b, err := New("test")
	require.NoError(t, err)

	var (
		errRun error
		wg     sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errRun = b.Run(ctx, WithUnordered(&testServiceBackoff{"s1"}))
	}()

	time.Sleep(400 * time.Millisecond)
	cancel()
	wg.Wait()

	require.NoError(t, errRun)

	require.Equal(t, int32(3), atomic.LoadInt32(&readyBackoff))
}
