package bux

import (
	"runtime/debug"
	"strings"

	"github.com/rs/zerolog"
)

func recoverAndLog(log *zerolog.Logger) {
	if err := recover(); err != nil {
		log.Error().Msgf(
			"panic: %v - stack trace: %v", err,
			strings.ReplaceAll(string(debug.Stack()), "\n", ""),
		)
	}
}
