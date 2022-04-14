package datastore

import (
	"context"
	"fmt"
	"time"

	zlogger "github.com/mrz1836/go-logger"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// newBasicLogger will return a basic logger interface
func newBasicLogger(debugging bool) logger.Interface {
	logLevel := logger.Warn
	if debugging {
		logLevel = logger.Info
	}
	return &basicLogger{LogLevel: logLevel}
}

// basicLogger is a basic implementation of the logger interface if no custom logger is provided
type basicLogger struct {
	LogLevel logger.LogLevel
}

// LogMode log mode
func (l *basicLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info print information
func (l *basicLogger) Info(_ context.Context, message string, params ...interface{}) {
	if l.LogLevel >= logger.Info {
		displayLog(zlogger.INFO, message, params...)
	}
}

// Warn print warn messages
func (l *basicLogger) Warn(_ context.Context, message string, params ...interface{}) {
	if l.LogLevel >= logger.Warn {
		displayLog(zlogger.WARN, message, params...)
	}
}

// Error print error messages
func (l *basicLogger) Error(_ context.Context, message string, params ...interface{}) {
	if l.LogLevel >= logger.Error {
		displayLog(zlogger.ERROR, message, params...)
	}
}

// Trace print sql message (Gorm Specific)
func (l *basicLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel >= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	level := zlogger.DEBUG

	params := []zlogger.KeyValue{
		zlogger.MakeParameter("executing_file", utils.FileWithLineNum()),
		zlogger.MakeParameter("elapsed", fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6)),
		zlogger.MakeParameter("rows", rows),
		zlogger.MakeParameter("sql", sql),
	}
	if err != nil {
		params = append(params, zlogger.MakeParameter("error_message", err.Error()))
		level = zlogger.ERROR
	}
	zlogger.Data(2, level, "sql trace", params...)
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
