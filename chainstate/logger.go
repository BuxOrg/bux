package chainstate

import (
	"context"
	"fmt"

	zlogger "github.com/mrz1836/go-logger"
	"gorm.io/gorm/logger"
)

// newBasicLogger will return a basic logger interface
func newBasicLogger(debugging bool) Logger {
	logLevel := logger.Warn
	if debugging {
		logLevel = logger.Info
	}
	return &basicLogger{LogLevel: logLevel}
}

// basicLogger is a basic logging implementation
type basicLogger struct {
	LogLevel logger.LogLevel
}

// LogMode log mode
func (l *basicLogger) LogMode(level logger.LogLevel) Logger {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info print information
func (l *basicLogger) Info(_ context.Context, message string, params ...interface{}) {
	if l.LogLevel <= logger.Info {
		displayLog(zlogger.INFO, message, params...)
	}
}

// displayLog will display a log using logger
func displayLog(level zlogger.LogLevel, message string, params ...interface{}) {
	var keyValues []zlogger.KeyValue
	if len(params) > 0 {
		for index, val := range params {
			keyValues = append(keyValues, zlogger.MakeParameter(fmt.Sprintf("index_%d", index), val))
		}
	}
	zlogger.Data(2, level, message, keyValues...)
}
