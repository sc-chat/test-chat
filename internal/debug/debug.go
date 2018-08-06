package debug

import "log"

// Logger interface
type Logger interface {
	Debug(layout string, args ...interface{})
}

// NopLog provides logger that does nothing
type NopLog struct{}

// Debug method does nothing
func (l *NopLog) Debug(layout string, args ...interface{}) {}

// Log provides basic stdout logger
type Log struct{}

// Debug writes output to stdout
func (l *Log) Debug(layout string, args ...interface{}) {
	log.Printf(layout, args...)
}

// NewLogger returns logger
func NewLogger(debug bool) Logger {
	if !debug {
		return new(NopLog)
	}

	return new(Log)
}
