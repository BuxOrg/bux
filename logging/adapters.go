package logging

import (
	"context"
	"github.com/mrz1836/go-logger"
	"github.com/rs/zerolog"
	"time"
)

type GormLoggerAdapter struct {
	Logger *zerolog.Logger
}

func (a *GormLoggerAdapter) Error(context.Context, string, ...interface{}) {
	a.Logger.Debug().Msg("test")
	a.Logger.Error().Msg("test")
}

func (a *GormLoggerAdapter) GetMode() logger.GormLogLevel {

}

func (a *GormLoggerAdapter) GetStackLevel() int {

}

func (a *GormLoggerAdapter) Info(context.Context, string, ...interface{}) {

}

func (a *GormLoggerAdapter) SetMode(logger.GormLogLevel) logger.GormLoggerInterface {

}

func (a *GormLoggerAdapter) SetStackLevel(level int) {

}

func (a *GormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {

}

func (a *GormLoggerAdapter) Warn(context.Context, string, ...interface{}) {

}
