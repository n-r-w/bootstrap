package executor

import (
	"context"
	"testing"
	"time"

	"github.com/n-r-w/bootstrap"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func TestService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockIExecutor(ctrl)

	var (
		executeCalled int
		testTime      = time.Second
		tickCount     = 10
		tickInterval  = testTime / time.Duration(tickCount)
	)

	mock.EXPECT().Execute(gomock.Any()).Do(func(ctx context.Context) {
		executeCalled++
	}).Return(nil).Times(tickCount)

	mock.EXPECT().StopExecutor(gomock.Any()).Return(nil).Times(1)

	service, err := New("testService", mock, tickInterval)
	require.NoError(t, err)
	require.Equal(t, bootstrap.Info{Name: "testService"}, service.Info())

	require.NoError(t, service.Start(context.Background()))

	time.Sleep(testTime)
	require.NoError(t, service.Stop(context.Background()))
	require.Equal(t, tickCount, executeCalled)

	// check that the service doesn't execute again after stop
	time.Sleep(tickInterval * 2)
	require.Equal(t, tickCount, executeCalled)
}
