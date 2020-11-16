package log

import (
	"context"
	"fmt"
	"time"
)

type Logger struct {
	outputLevel uint
}

type LogWriter struct {
	Info func(format string, a ...interface{})
}

type contextKey string

var (
	consoleWriter              = LogWriter{Info: info}
	noneWriter                 = LogWriter{Info: func(format string, a ...interface{}) {}}
	logLevelKey     contextKey = "ll"
	defaultLogLevel uint       = 2
)

// FromContext returns logger from context
func FromContext(ctx context.Context) *Logger {
	logLevel := defaultLogLevel
	if l := ctx.Value(logLevelKey); l != nil {
		if v, ok := l.(uint); ok {
			logLevel = v
		}
	}

	return &Logger{outputLevel: logLevel}
}

// WithLogLevel create a context with logging level
func WithLogLevel(ctx context.Context, logLevel uint) context.Context {
	return context.WithValue(ctx, logLevelKey, logLevel)
}

// V returns log writter at log level
func (l *Logger) V(level uint) *LogWriter {
	if level <= l.outputLevel {
		return &consoleWriter
	}
	return &noneWriter
}

func info(format string, a ...interface{}) {
	fmt.Printf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), fmt.Sprintf(format, a...))
}
