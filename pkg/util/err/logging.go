package errutil

import (
	"context"
	"fmt"
	"log"

	"go.uber.org/zap"
)

var (
	verbose bool
	// L is the global logger.
	L = Logger{zap.NewNop()}
)

type key int

const (
	logKey  key = iota
	logFile     = "/tmp/artifacts.log"
)

// Init initializes logging, based on we want verbose, or simple logs. Nonverbose logs go to a
// temporary file only, and the user get a nice message about everything from Info to Error.
func Init(v bool) {
	verbose = v
	var c zap.Config
	if v {
		c = zap.NewDevelopmentConfig()
		c.OutputPaths = []string{"stderr", logFile}
		c.ErrorOutputPaths = []string{"stderr", logFile}
	} else {
		c = zap.NewProductionConfig()
		c.OutputPaths = []string{logFile}
		c.ErrorOutputPaths = []string{logFile}
	}
	l, err := c.Build()
	if err != nil {
		panic(fmt.Errorf("failed to initialize logger: %s", err.Error()))
	}
	L = Logger{l}
}

// Logger is a wrapper for zap logger that may have custom functions on it.
type Logger struct {
	*zap.Logger
}

// Error wraps printing a nice message to the user, and logging the error to the zap logger.
func (l Logger) Error(msg string, fields ...zap.Field) {
	log.Println(msg)
	l.Logger.Error(msg, fields...)
}

// Warn wraps printing a nice message to the user, and logging the warning to the zap logger.
func (l Logger) Warn(msg string, fields ...zap.Field) {
	log.Println(msg)
	l.Logger.Warn(msg, fields...)
}

// Info wraps printing a nice message to the user, and logging the information to the zap logger.
func (l Logger) Info(msg string, fields ...zap.Field) {
	log.Println(msg)
	l.Logger.Info(msg, fields...)
}

// Debug wraps printing a nice message to the user, and logging the debug to the zap logger.
func (l Logger) Debug(msg string, fields ...zap.Field) {
	if verbose {
		log.Println(msg)
		l.Logger.Debug(msg, fields...)
	}
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
