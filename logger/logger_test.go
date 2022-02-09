package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	glogger "gorm.io/gorm/logger"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("basic logger", func(t *testing.T) {
		l := NewLogger(true)
		require.NotNil(t, l)

		l = NewLogger(false)
		require.NotNil(t, l)
	})
}

func TestBasicLogger_LogMode(t *testing.T) {
	t.Parallel()

	t.Run("new mode", func(t *testing.T) {
		l := NewLogger(true)
		require.NotNil(t, l)

		l2 := l.LogMode(glogger.Info)
		require.NotNil(t, l2)
	})
}
