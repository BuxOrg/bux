package bux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// todo: finish unit tests!

// TestNew will test the method New()
func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("Get opts", func(t *testing.T) {
		opt := New()
		assert.IsType(t, *new(ModelOps), opt)
	})

	t.Run("apply opts", func(t *testing.T) {
		opt := New()
		m := new(Model)
		m.SetOptions(opt)
		assert.Equal(t, true, m.IsNew())
	})
}

// TestWithMetadata will test the method WithMetadata()
func TestWithMetadata(t *testing.T) {
	t.Parallel()

	t.Run("Get opts", func(t *testing.T) {
		opt := WithMetadata("key", "value")
		assert.IsType(t, *new(ModelOps), opt)
	})

	t.Run("apply opts", func(t *testing.T) {
		opt := WithMetadata("key", "value")
		m := new(Model)
		m.SetOptions(opt)
		assert.Equal(t, "value", m.Metadata["key"])
	})
}

// TestWithClient will test the method WithClient()
func TestWithClient(t *testing.T) {
	// finish test
}

// TestWithMetadatas will test the method WithMetadatas()
func TestWithMetadatas(t *testing.T) {
	// finish test
}
