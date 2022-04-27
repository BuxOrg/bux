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
	Error(context.Context, string, ...interface{})
	GetMode() LogLevel
	GetStackLevel() int
	Info(context.Context, string, ...interface{})
	SetMode(LogLevel) Interface
	SetStackLevel(level int)
	Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
	Warn(context.Context, string, ...interface{})
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

const slowQueryThreshold = 5 * time.Second

// NewLogger will return a basic logger interface
func NewLogger(debugging bool, stackLevel int) Interface {
	logLevel := Warn
	if debugging {
		logLevel = Info
	}
	return &basicLogger{
		logLevel:   logLevel,
		stackLevel: stackLevel,
	}
}

// basicLogger is a basic implementation of the logger interface if no custom logger is provided
type basicLogger struct {
	logLevel   LogLevel // Log level (info, error, etc)
	stackLevel int      // How many files/functions to traverse upwards to record the file/line
}

// SetMode will set the log mode
func (l *basicLogger) SetMode(level LogLevel) Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// SetStackLevel will set the stack level
func (l *basicLogger) SetStackLevel(level int) {
	l.stackLevel = level
}

// GetStackLevel will get the current stack level
func (l *basicLogger) GetStackLevel() int {
	return l.stackLevel
}

// GetMode will get the log mode
func (l *basicLogger) GetMode() LogLevel {
	return l.logLevel
}

// Info print information
func (l *basicLogger) Info(_ context.Context, message string, params ...interface{}) {
	if l.logLevel >= Info {
		displayLog(zlogger.INFO, l.stackLevel, message, params...)
	}
}

// Warn print warn messages
func (l *basicLogger) Warn(_ context.Context, message string, params ...interface{}) {
	if l.logLevel >= Warn {
		displayLog(zlogger.WARN, l.stackLevel, message, params...)
	}
}

// Error print error messages
func (l *basicLogger) Error(_ context.Context, message string, params ...interface{}) {
	if l.logLevel >= Error {
		displayLog(zlogger.ERROR, l.stackLevel, message, params...)
	}
}

// Trace is for GORM/SQL tracing from datastore
func (l *basicLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.logLevel >= Error && (!errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		zlogger.Data(l.stackLevel, zlogger.ERROR,
			"error executing query",
			zlogger.MakeParameter("file", utils.FileWithLineNum()),
			zlogger.MakeParameter("error", err.Error()),
			zlogger.MakeParameter("duration", fmt.Sprintf("[%.3fms]", float64(elapsed.Nanoseconds())/1e6)),
			zlogger.MakeParameter("rows", rows),
			zlogger.MakeParameter("sql", sql),
		)
	case elapsed > slowQueryThreshold && l.logLevel >= Warn:
		sql, rows := fc()
		zlogger.Data(l.stackLevel, zlogger.WARN,
			"warning executing query",
			zlogger.MakeParameter("file", utils.FileWithLineNum()),
			zlogger.MakeParameter("slow_log", fmt.Sprintf("SLOW SQL >= %v", slowQueryThreshold)),
			zlogger.MakeParameter("duration", fmt.Sprintf("[%.3fms]", float64(elapsed.Nanoseconds())/1e6)),
			zlogger.MakeParameter("rows", rows),
			zlogger.MakeParameter("sql", sql),
		)
	case l.logLevel == Info:
		sql, rows := fc()
		zlogger.Data(l.stackLevel, zlogger.WARN,
			"executing sql query",
			zlogger.MakeParameter("file", utils.FileWithLineNum()),
			zlogger.MakeParameter("duration", fmt.Sprintf("[%.3fms]", float64(elapsed.Nanoseconds())/1e6)),
			zlogger.MakeParameter("rows", rows),
			zlogger.MakeParameter("sql", sql),
		)
	}
}

// displayLog will display a log using logger
func displayLog(level zlogger.LogLevel, stackLevel int, message string, params ...interface{}) {
	var keyValues []zlogger.KeyValue
	if len(params) > 0 {
		for index, val := range params {
			keyValues = append(keyValues, zlogger.MakeParameter(fmt.Sprintf("param_%d", index), val))
		}
	}
	zlogger.Data(stackLevel, level, message, keyValues...)
}
