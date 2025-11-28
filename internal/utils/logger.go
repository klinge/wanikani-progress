package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the global logger with the specified log level
func Init(level string) *logrus.Logger {
	log = logrus.New()
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Parse and set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	return log
}

// Get returns the global logger instance
func Get() *logrus.Logger {
	if log == nil {
		return Init("info")
	}
	return log
}
