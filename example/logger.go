package main

import (
	"context"
	"fmt"
	"log/slog"
)

type customLogger struct {
	l *slog.Logger
}

func newCustomLogger() *customLogger {
	return &customLogger{
		l: slog.Default(),
	}
}

func (c *customLogger) Debugf(ctx context.Context, format string, args ...any) {
	c.l.DebugContext(ctx, format, slog.String("message", fmt.Sprintf(format, args...)))
}

func (c *customLogger) Infof(ctx context.Context, format string, args ...any) {
	c.l.InfoContext(ctx, format, slog.String("message", fmt.Sprintf(format, args...)))
}

func (c *customLogger) Warningf(ctx context.Context, format string, args ...any) {
	c.l.WarnContext(ctx, format, slog.String("message", fmt.Sprintf(format, args...)))
}

func (c *customLogger) Errorf(ctx context.Context, format string, args ...any) {
	c.l.ErrorContext(ctx, format, slog.String("message", fmt.Sprintf(format, args...)))
}
