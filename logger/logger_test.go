package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("basic logger", func(t *testing.T) {
		l := NewLogger(true, 3)
		require.NotNil(t, l)

		l = NewLogger(false, 3)
		require.NotNil(t, l)
	})
}

func TestBasicLogger_LogMode(t *testing.T) {
	t.Parallel()

	t.Run("new mode", func(t *testing.T) {
		l := NewLogger(true, 3)
		require.NotNil(t, l)

		l2 := l.SetMode(Info)
		require.NotNil(t, l2)

		mode := l.GetMode()
		assert.Equal(t, Info, mode)
	})
}

func TestBasicLogger_SetStackLevel(t *testing.T) {
	t.Parallel()

	t.Run("set/get level", func(t *testing.T) {
		l := NewLogger(true, 3)
		require.NotNil(t, l)

		l.SetStackLevel(3)

		level := l.GetStackLevel()
		assert.Equal(t, 3, level)
	})
}
