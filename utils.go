package bux

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	zLogger "github.com/mrz1836/go-logger"
)

func recoverAndLog(ctx context.Context, log zLogger.GormLoggerInterface) {
	if err := recover(); err != nil {
		log.Error(ctx,
			fmt.Sprintf(
				"panic: %v - stack trace: %v", err,
				strings.ReplaceAll(string(debug.Stack()), "\n", ""),
			),
		)
	}
}
