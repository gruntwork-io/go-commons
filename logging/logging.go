package logging

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var globalLogLevel = logrus.InfoLevel
var globalLogLevelLock = sync.Mutex{}

var globalLogFormatter = "text"
var globalLogFormatterLock = sync.Mutex{}

func GetProjectLogger() *logrus.Entry {
	logger := GetLogger("", "")
	return logger.WithField("name", "go-commons")
}

// Create a new logger with the given name
func GetLogger(name string, version string) *logrus.Entry {
	logger := logrus.New()

	logger.Level = globalLogLevel

	if globalLogFormatter == "json" {
		logger.Formatter = &logrus.JSONFormatter{}

	} else {
		logger.Formatter = &logrus.TextFormatter{
			FullTimestamp: true,
		}
	}
	return logger.WithField("binary", name).WithField("version", version)

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

// Set the log format. Note: this ONLY affects loggers created using the GetLogger function AFTER this function has been
// called. Therefore, you need to call this as early in the life of your CLI app as possible!
func SetGlobalLogFormatter(formatter string) {
	defer globalLogFormatterLock.Unlock()
	globalLogFormatterLock.Lock()
	globalLogFormatter = formatter
}
