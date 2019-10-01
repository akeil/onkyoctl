package onkyoctl

import (
	"io/ioutil"
	"log"
	"os"
)

// LogLevel is the type for log levels.
type LogLevel int

const (
	// Debug log level
	Debug LogLevel = iota
	// Info log level
	Info
	// Warning log level
	Warning
	// Error log level
	Error
	// NoLog means no messages are logged, regardless of severity
	NoLog
)

// Logger is the interface used for logging.
type Logger interface {
	Debug(msg string, v ...interface{})
	Info(msg string, v ...interface{})
	Warning(msg string, v ...interface{})
	Error(msg string, v ...interface{})
}

// NewLogger returns a Logger with the given log level.
func NewLogger(level LogLevel) Logger {
	flags := log.Ldate | log.Ltime | log.LUTC
	l := &basicLogger{
		debug:   log.New(ioutil.Discard, "D ", flags),
		info:    log.New(ioutil.Discard, "I ", flags),
		warning: log.New(ioutil.Discard, "W ", flags),
		error:   log.New(ioutil.Discard, "E ", flags),
	}

	if level <= Debug {
		l.debug.SetOutput(os.Stderr)
	}

	if level <= Info {
		l.info.SetOutput(os.Stderr)
	}

	if level <= Warning {
		l.warning.SetOutput(os.Stderr)
	}

	if level <= Error {
		l.error.SetOutput(os.Stderr)
	}

	return l
}

type basicLogger struct {
	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
	error   *log.Logger
}

func (l *basicLogger) Debug(msg string, v ...interface{}) {
	l.debug.Printf(msg, v...)
}

func (l *basicLogger) Info(msg string, v ...interface{}) {
	l.info.Printf(msg, v...)
}

func (l *basicLogger) Warning(msg string, v ...interface{}) {
	l.warning.Printf(msg, v...)
}

func (l *basicLogger) Error(msg string, v ...interface{}) {
	l.error.Printf(msg, v...)
}
