// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go
//
// Generated by this command:
//
//	mockgen -source interface.go -destination interface_mock.go -package bootstrap
//

// Package bootstrap is a generated GoMock package.
package bootstrap

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockIService is a mock of IService interface.
type MockIService struct {
	ctrl     *gomock.Controller
	recorder *MockIServiceMockRecorder
}

// MockIServiceMockRecorder is the mock recorder for MockIService.
type MockIServiceMockRecorder struct {
	mock *MockIService
}

// NewMockIService creates a new mock instance.
func NewMockIService(ctrl *gomock.Controller) *MockIService {
	mock := &MockIService{ctrl: ctrl}
	mock.recorder = &MockIServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIService) EXPECT() *MockIServiceMockRecorder {
	return m.recorder
}

// Info mocks base method.
func (m *MockIService) Info() Info {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Info")
	ret0, _ := ret[0].(Info)
	return ret0
}

// Info indicates an expected call of Info.
func (mr *MockIServiceMockRecorder) Info() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockIService)(nil).Info))
}

// Start mocks base method.
func (m *MockIService) Start(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockIServiceMockRecorder) Start(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockIService)(nil).Start), ctx)
}

// Stop mocks base method.
func (m *MockIService) Stop(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockIServiceMockRecorder) Stop(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockIService)(nil).Stop), ctx)
}

// MockIHealthChecher is a mock of IHealthChecher interface.
type MockIHealthChecher struct {
	ctrl     *gomock.Controller
	recorder *MockIHealthChecherMockRecorder
}

// MockIHealthChecherMockRecorder is the mock recorder for MockIHealthChecher.
type MockIHealthChecherMockRecorder struct {
	mock *MockIHealthChecher
}

// NewMockIHealthChecher creates a new mock instance.
func NewMockIHealthChecher(ctrl *gomock.Controller) *MockIHealthChecher {
	mock := &MockIHealthChecher{ctrl: ctrl}
	mock.recorder = &MockIHealthChecherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIHealthChecher) EXPECT() *MockIHealthChecherMockRecorder {
	return m.recorder
}

// AddCheckErrorHandler mocks base method.
func (m *MockIHealthChecher) AddCheckErrorHandler(handler func(string, error)) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddCheckErrorHandler", handler)
}

// AddCheckErrorHandler indicates an expected call of AddCheckErrorHandler.
func (mr *MockIHealthChecherMockRecorder) AddCheckErrorHandler(handler any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddCheckErrorHandler", reflect.TypeOf((*MockIHealthChecher)(nil).AddCheckErrorHandler), handler)
}

// AddLivenessCheck mocks base method.
func (m *MockIHealthChecher) AddLivenessCheck(name string, check func() error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddLivenessCheck", name, check)
}

// AddLivenessCheck indicates an expected call of AddLivenessCheck.
func (mr *MockIHealthChecherMockRecorder) AddLivenessCheck(name, check any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddLivenessCheck", reflect.TypeOf((*MockIHealthChecher)(nil).AddLivenessCheck), name, check)
}

// AddReadinessCheck mocks base method.
func (m *MockIHealthChecher) AddReadinessCheck(name string, check func() error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddReadinessCheck", name, check)
}

// AddReadinessCheck indicates an expected call of AddReadinessCheck.
func (mr *MockIHealthChecherMockRecorder) AddReadinessCheck(name, check any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddReadinessCheck", reflect.TypeOf((*MockIHealthChecher)(nil).AddReadinessCheck), name, check)
}

// MockILogger is a mock of ILogger interface.
type MockILogger struct {
	ctrl     *gomock.Controller
	recorder *MockILoggerMockRecorder
}

// MockILoggerMockRecorder is the mock recorder for MockILogger.
type MockILoggerMockRecorder struct {
	mock *MockILogger
}

// NewMockILogger creates a new mock instance.
func NewMockILogger(ctrl *gomock.Controller) *MockILogger {
	mock := &MockILogger{ctrl: ctrl}
	mock.recorder = &MockILoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockILogger) EXPECT() *MockILoggerMockRecorder {
	return m.recorder
}

// Debugf mocks base method.
func (m *MockILogger) Debugf(ctx context.Context, format string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, format}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debugf", varargs...)
}

// Debugf indicates an expected call of Debugf.
func (mr *MockILoggerMockRecorder) Debugf(ctx, format any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, format}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debugf", reflect.TypeOf((*MockILogger)(nil).Debugf), varargs...)
}

// Errorf mocks base method.
func (m *MockILogger) Errorf(ctx context.Context, format string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, format}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Errorf", varargs...)
}

// Errorf indicates an expected call of Errorf.
func (mr *MockILoggerMockRecorder) Errorf(ctx, format any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, format}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errorf", reflect.TypeOf((*MockILogger)(nil).Errorf), varargs...)
}

// Infof mocks base method.
func (m *MockILogger) Infof(ctx context.Context, format string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, format}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Infof", varargs...)
}

// Infof indicates an expected call of Infof.
func (mr *MockILoggerMockRecorder) Infof(ctx, format any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, format}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Infof", reflect.TypeOf((*MockILogger)(nil).Infof), varargs...)
}

// Warningf mocks base method.
func (m *MockILogger) Warningf(ctx context.Context, format string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, format}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warningf", varargs...)
}

// Warningf indicates an expected call of Warningf.
func (mr *MockILoggerMockRecorder) Warningf(ctx, format any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, format}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warningf", reflect.TypeOf((*MockILogger)(nil).Warningf), varargs...)
}
