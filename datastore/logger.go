package datastore

import (
	"github.com/BuxOrg/bux/logger"
	glogger "gorm.io/gorm/logger"
)

// DatabaseLogWrapper is a special wrapper for the GORM logger
type DatabaseLogWrapper struct {
	logger.Interface
}

// LogMode will set the log level/mode
func (d *DatabaseLogWrapper) LogMode(level glogger.LogLevel) glogger.Interface {
	newLogger := *d
	if level == glogger.Info {
		newLogger.SetMode(logger.Info)
	} else if level == glogger.Warn {
		newLogger.SetMode(logger.Warn)
	} else if level == glogger.Error {
		newLogger.SetMode(logger.Error)
	} else if level == glogger.Silent {
		newLogger.SetMode(logger.Silent)
	}

	return &newLogger
}
