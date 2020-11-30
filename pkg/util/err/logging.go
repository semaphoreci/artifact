package errutil

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/semaphoreci/artifact/config"
	"go.uber.org/zap"
)

// L is the global logger.
var L Logger

type key int

const (
	logKey key = iota
)

func init() {
	if strings.HasSuffix(os.Args[0], ".test") { // testing
		L = Logger{zap.NewNop()}
	} else {
		var err error
		var l *zap.Logger
		if config.LogLevel > config.LogLvlDebug {
			l, err = zap.NewProduction()
		} else {
			l, err = zap.NewDevelopment()
		}
		if err != nil {
			panic(fmt.Errorf("failed to initialize logger: %s", err.Error()))
		}
		L = Logger{l}
	}
}

// Logger is a wrapper for zap logger that may have custom functions on it.
type Logger struct {
	*zap.Logger
}

// CreateContext returns a new logger with the given fields tagged to the logger, and the
// context containing this logger.
func CreateContext(ctx context.Context, fields ...zap.Field) (context.Context, *zap.Logger) {
	l := L.With(fields...)
	return context.WithValue(ctx, logKey, l), l
}

// CreateContextNop adds a noop logger to the context.
func CreateContextNop(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, logKey, Logger{zap.NewNop()})
}

// WithContext returns a logger related to the given context.
func WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return L
	}

	if ctxLogger, ok := ctx.Value(logKey).(Logger); ok {
		return ctxLogger
	}
	return L
}
