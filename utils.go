package bux

import (
	"context"
	"github.com/rs/zerolog"
	"runtime/debug"
	"strings"
)

func recoverAndLog(ctx context.Context, log *zerolog.Logger) {
	if err := recover(); err != nil {
		log.Error().Msgf(
			"panic: %v - stack trace: %v", err,
			strings.ReplaceAll(string(debug.Stack()), "\n", ""),
		)
	}
}
