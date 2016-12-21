package logging

import (
	"github.com/Sirupsen/logrus"
	"fmt"
	"sync"
)

const loggerNameField = "name"

var globalLogLevel = logrus.InfoLevel
var globalLogLevelLock = sync.Mutex{}

// Create a new logger with the given name
func GetLogger(name string) *logrus.Entry {
	logger := logrus.New()
	logger.Level = globalLogLevel
	if name != "" {
		return logger.WithField(loggerNameField, name)
	} else {
		return logger.WithFields(make(logrus.Fields))
	}
}

// Set the log level. Note: this ONLY affects loggers created using the GetLogger function AFTER this function has been
// called. Therefore, you need to call this as early in the life of your CLI app as possible!
func SetGlobalLogLevel(level logrus.Level) {
	// We need to lock here as this function may be called from multiple threads concurrently (e.g. especially at
	// test time)
	defer globalLogLevelLock.Unlock()
	globalLogLevelLock.Lock()

	globalLogLevel = level
}

// Return the name of the given logger as set by the GetLogger function or an empty string if this logger has no name
func GetLoggerName(logger *logrus.Entry) string {
	name, hasName := logger.Data[loggerNameField]
	if hasName {
		return fmt.Sprintf("%v", name)
	} else {
		return ""
	}
}
