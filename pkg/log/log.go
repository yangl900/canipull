package log

import (
	"fmt"
	"time"
)

type Logger struct {
	outputLevel uint
}

type LogWriter struct {
	Info func(format string, a ...interface{})
}

var (
	consoleWriter = LogWriter{Info: info}
	noneWriter    = LogWriter{Info: func(format string, a ...interface{}) {}}
)

func NewLogger(outputLevel uint) *Logger {
	return &Logger{outputLevel: outputLevel}
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
