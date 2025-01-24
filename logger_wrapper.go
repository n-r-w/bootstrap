package bootstrap

import (
	"context"
)

type loggerWrapper struct {
	logger ILogger
}

func (l *loggerWrapper) Debugf(ctx context.Context, format string, args ...any) {
	if l.logger != nil {
		l.logger.Debugf(ctx, format, args...)
	}
}

func (l *loggerWrapper) Infof(ctx context.Context, format string, args ...any) {
	if l.logger != nil {
		l.logger.Infof(ctx, format, args...)
	}
}

func (l *loggerWrapper) Warningf(ctx context.Context, format string, args ...any) {
	if l.logger != nil {
		l.logger.Warningf(ctx, format, args...)
	}
}

func (l *loggerWrapper) Errorf(ctx context.Context, format string, args ...any) {
	if l.logger != nil {
		l.logger.Errorf(ctx, format, args...)
	}
}
