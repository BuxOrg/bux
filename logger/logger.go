package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	zlogger "github.com/mrz1836/go-logger"
	"gorm.io/gorm"
	"gorm.io/gorm/utils"
)

// Interface is a logger interface
type Interface interface {
	SetMode(LogLevel) Interface
	GetMode() LogLevel
	Info(context.Context, string, ...interface{})
	Warn(context.Context, string, ...interface{})
	Error(context.Context, string, ...interface{})
	Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
}

// LogLevel is the log level
type LogLevel int

const (
	// Silent silent log level
	Silent LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
)

const slowThreshold = 5 * time.Second

// NewLogger will return a basic logger interface
func NewLogger(debugging bool) Interface {
	logLevel := Warn
	if debugging {
		logLevel = Info
	}
	return &basicLogger{logLevel: logLevel}
}

// basicLogger is a basic implementation of the logger interface if no custom logger is provided
type basicLogger struct {
	logLevel LogLevel
}

// SetMode will set the log mode
func (l *basicLogger) SetMode(level LogLevel) Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// GetMode will get the log mode
func (l *basicLogger) GetMode() LogLevel {
	return l.logLevel
}

// Info print information
func (l *basicLogger) Info(_ context.Context, message string, params ...interface{}) {
	if l.logLevel >= Info {
		displayLog(zlogger.INFO, message, params...)
	}
}

// Warn print warn messages
func (l *basicLogger) Warn(_ context.Context, message string, params ...interface{}) {
	if l.logLevel >= Warn {
		displayLog(zlogger.WARN, message, params...)
	}
}

// Error print error messages
func (l *basicLogger) Error(_ context.Context, message string, params ...interface{}) {
	if l.logLevel >= Error {
		displayLog(zlogger.ERROR, message, params...)
	}
}

// Trace is for GORM and SQL tracing
func (l *basicLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.logLevel >= Error && (!errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		if rows == -1 {
			zlogger.Data(4, zlogger.ERROR,
				fmt.Sprintf("%s %s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql),
			)
		} else {
			zlogger.Data(4, zlogger.ERROR,
				fmt.Sprintf("%s %s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql),
			)
		}
	case elapsed > slowThreshold && l.logLevel >= Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", slowThreshold)
		if rows == -1 {
			zlogger.Data(4, zlogger.WARN,
				fmt.Sprintf("%s %s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql),
			)
		} else {
			zlogger.Data(4, zlogger.WARN,
				fmt.Sprintf("%s %s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql),
			)
		}
	case l.logLevel == Info:
		sql, rows := fc()
		if rows == -1 {
			zlogger.Data(4, zlogger.INFO,
				fmt.Sprintf("%s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql),
			)
		} else {
			zlogger.Data(4, zlogger.INFO,
				fmt.Sprintf("%s\n[%.3fms] [rows:%v] %s", utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql),
			)
		}
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
	zlogger.Data(4, level, message, keyValues...)
}
