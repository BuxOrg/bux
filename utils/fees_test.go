package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetInputSizeForType will test the method GetInputSizeForType()
func TestGetInputSizeForType(t *testing.T) {
	t.Parallel()

	t.Run("valid input type", func(t *testing.T) {
		assert.Equal(t, uint64(148), GetInputSizeForType(ScriptTypePubKeyHash))
	})

	t.Run("unknown input type", func(t *testing.T) {
		assert.Equal(t, uint64(500), GetInputSizeForType("unknown"))
	})
}

// TestGetOutputSizeForType will test the method GetOutputSizeForType()
func TestGetOutputSizeForType(t *testing.T) {
	t.Parallel()

	t.Run("valid output type", func(t *testing.T) {
		assert.Equal(t, uint64(34), GetOutputSize("76a914a7bf13994cb80a6c17ca3624cae128bf1ff4c57b88ac"))
	})

	t.Run("unknown input type", func(t *testing.T) {
		assert.Equal(t, uint64(500), GetOutputSize(""))
	})
}
