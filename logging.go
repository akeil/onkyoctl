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

var flags = log.Ldate | log.Ltime | log.LUTC
var (
	debugLogger   = log.New(ioutil.Discard, "D ", flags)
	infoLogger    = log.New(ioutil.Discard, "I ", flags)
	warningLogger = log.New(os.Stderr, "W ", flags)
	errorLogger   = log.New(os.Stderr, "E ", flags)
)

// SetLogLevel sets the logging verbosity.
func SetLogLevel(level LogLevel) {
	if level <= Debug {
		debugLogger.SetOutput(os.Stderr)
	} else {
		debugLogger.SetOutput(ioutil.Discard)
	}

	if level <= Info {
		infoLogger.SetOutput(os.Stderr)
	} else {
		infoLogger.SetOutput(ioutil.Discard)
	}

	if level <= Warning {
		warningLogger.SetOutput(os.Stderr)
	} else {
		warningLogger.SetOutput(ioutil.Discard)
	}

	if level <= Error {
		errorLogger.SetOutput(os.Stderr)
	} else {
		errorLogger.SetOutput(ioutil.Discard)
	}
}

// LogDebug emits a log message with level DEBUG.
func logDebug(message string, v ...interface{}) {
	debugLogger.Printf(message, v...)
}

// LogInfo emits a log message with level INFO.
func logInfo(message string, v ...interface{}) {
	infoLogger.Printf(message, v...)
}

// LogWarning emits a log message with level WARNING.
func logWarning(message string, v ...interface{}) {
	warningLogger.Printf(message, v...)
}

// LogError emits a log message with level ERROR.
func logError(message string, v ...interface{}) {
	errorLogger.Printf(message, v...)
}
