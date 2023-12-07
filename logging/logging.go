package logging

import (
	"github.com/rs/zerolog"
	"go.elastic.co/ecszerolog"
	"os"

	"context"
)

func GetDefaultLogger() *zerolog.Logger {
	logger := ecszerolog.New(os.Stdout, ecszerolog.Level(zerolog.DebugLevel)).
		With().
		Timestamp().
		Caller().
		Str("application", "bux-default").
		Logger()

	return &logger
}

type Logger struct {
	*zerolog.Logger
}

func (l *Logger) Error(ctx context.Context, s string, i ...interface{}) {

}
