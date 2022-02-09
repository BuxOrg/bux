package chainstate

import (
	"context"
	"fmt"

	zlogger "github.com/mrz1836/go-logger"
)

// newLogger will return a basic logger interface
func newLogger() Logger {
	return &basicLogger{}
}

// basicLogger is a basic logging implementation
type basicLogger struct{}

// Info print information
func (l *basicLogger) Info(_ context.Context, message string, params ...interface{}) {
	displayLog(zlogger.INFO, message, params...)
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
