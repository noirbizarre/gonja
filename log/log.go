package log

import (
	"fmt"
	"log"
	"strings"
)

type Logger interface {
	Trace(msg string, fields ...any)
	Error(msg string, fields ...any)
}

type defaultLogger struct{}

func (d defaultLogger) Trace(msg string, fields ...any) {
	d.log("TRACE", msg, fields...)
}

func (d defaultLogger) Error(msg string, fields ...any) {
	d.log("ERROR", msg, fields...)
}

func (d defaultLogger) log(level string, msg string, fields ...any) {
	if len(fields)%2 != 0 {
		panic("fields must have even number of elements")
	}

	var buf strings.Builder
	buf.WriteString("[" + level + "] " + msg + " ")
	for i := 0; i < len(fields); i += 2 {
		_, _ = fmt.Fprintf(&buf, " %v=%v", fields[i], fields[i+1])
	}

	log.Print(buf.String())
}

type Level int

const (
	InfoLevel Level = iota
	TraceLevel
)

var DefaultLogger Logger = &defaultLogger{}
var DefaultLevel Level = InfoLevel

func Trace(msg string, fields ...any) {
	if DefaultLevel >= TraceLevel {
		DefaultLogger.Trace(msg, fields...)
	}
}

func Error(msg string, fields ...any) {
	if DefaultLevel >= InfoLevel {
		DefaultLogger.Error(msg, fields...)
	}
}
