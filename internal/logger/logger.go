// Package logger provides central logging
package logger

import (
	log "github.com/sirupsen/logrus"
)

var logger *log.Logger // nolint:gochecknoglobals

// Get returns a global logger
func Get() *log.Logger {
	return logger
}

// Init sets the global logger
func Init(logLevel string) {
	if logger == nil {
		logLevelValue, err := log.ParseLevel(logLevel)
		if err != nil {
			logger.Panic("invalid log level, ", err)
		}

		logger = log.New()
		logger.SetReportCaller(true)
		logger.SetLevel(logLevelValue)
	}
}
