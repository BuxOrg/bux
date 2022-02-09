package datastore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEngine_String will test the method String()
func TestEngine_String(t *testing.T) {
	t.Run("valid name", func(t *testing.T) {
		assert.Equal(t, "empty", Empty.String())
		assert.Equal(t, "mongodb", MongoDB.String())
		assert.Equal(t, "mysql", MySQL.String())
		assert.Equal(t, "postgresql", PostgreSQL.String())
		assert.Equal(t, "sqlite", SQLite.String())
	})
}

// TestEngine_IsEmpty will test the method IsEmpty()
func TestEngine_IsEmpty(t *testing.T) {
	t.Run("actually empty", func(t *testing.T) {
		assert.Equal(t, true, Empty.IsEmpty())
	})

	t.Run("not empty", func(t *testing.T) {
		assert.Equal(t, false, MySQL.IsEmpty())
	})
}

// TestIsSQLEngine will test the method IsSQLEngine()
func TestIsSQLEngine(t *testing.T) {
	t.Run("test sql databases", func(t *testing.T) {
		assert.Equal(t, true, IsSQLEngine(MySQL))
		assert.Equal(t, true, IsSQLEngine(PostgreSQL))
		assert.Equal(t, true, IsSQLEngine(SQLite))
	})

	t.Run("test other databases", func(t *testing.T) {
		assert.Equal(t, false, IsSQLEngine(MongoDB))
		assert.Equal(t, false, IsSQLEngine(Empty))
	})
}
