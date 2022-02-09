package datastore

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// todo: finish unit tests!

// TestNewClient will test the method NewClient()
func TestNewClient(t *testing.T) {
	// finish test
}

// TestClient_Close will test the method Close()
func TestClient_Close(t *testing.T) {
	// finish test
}

// TestClient_IsDebug will test the method IsDebug()
func TestClient_IsDebug(t *testing.T) {
	t.Run("toggle debug", func(t *testing.T) {
		c, err := NewClient(context.Background(), WithDebugging())
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, true, c.IsDebug())

		c.Debug(false)

		assert.Equal(t, false, c.IsDebug())
	})

	// Attempt to remove a file created during the test
	t.Cleanup(func() {
		_ = os.Remove("datastore.db")
	})
}

// TestClient_Debug will test the method Debug()
func TestClient_Debug(t *testing.T) {
	t.Run("turn debug on", func(t *testing.T) {
		c, err := NewClient(context.Background())
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, false, c.IsDebug())

		c.Debug(true)

		assert.Equal(t, true, c.IsDebug())
	})

	// Attempt to remove a file created during the test
	t.Cleanup(func() {
		_ = os.Remove("datastore.db")
	})
}

// TestClient_DebugLog will test the method DebugLog()
func TestClient_DebugLog(t *testing.T) {
	t.Run("write debug log", func(t *testing.T) {
		c, err := NewClient(context.Background(), WithDebugging())
		require.NotNil(t, c)
		require.NoError(t, err)

		c.DebugLog("test message")
	})

	// Attempt to remove a file created during the test
	t.Cleanup(func() {
		_ = os.Remove("datastore.db")
	})
}

// TestClient_Engine will test the method Engine()
func TestClient_Engine(t *testing.T) {
	t.Run("[sqlite] - get engine", func(t *testing.T) {
		c, err := NewClient(context.Background(), WithSQLite(&SQLiteConfig{
			DatabasePath: "",
			Shared:       true,
		}))
		assert.NotNil(t, c)
		assert.NoError(t, err)
		assert.Equal(t, SQLite, c.Engine())
	})

	t.Run("[mongo] - failed to load", func(t *testing.T) {
		c, err := NewClient(context.Background(), WithMongo(&MongoDBConfig{
			DatabaseName: "test",
			Transactions: false,
			URI:          "",
		}))
		assert.Nil(t, c)
		assert.Error(t, err)
	})

	// todo: add MySQL, Postgresql and MongoDB
}

// TestClient_GetTableName will test the method GetTableName()
func TestClient_GetTableName(t *testing.T) {
	t.Run("table prefix", func(t *testing.T) {
		c, err := NewClient(context.Background(), WithDebugging(), WithSQLite(&SQLiteConfig{
			CommonConfig: CommonConfig{
				TablePrefix: testTablePrefix,
			},
			DatabasePath: "",
			Shared:       true,
		}))
		require.NotNil(t, c)
		require.NoError(t, err)

		tableName := c.GetTableName(testModelName)
		assert.Equal(t, testTablePrefix+"_"+testModelName, tableName)
	})

	t.Run("no table prefix", func(t *testing.T) {
		c, err := NewClient(context.Background(), WithDebugging(), WithSQLite(&SQLiteConfig{
			CommonConfig: CommonConfig{
				TablePrefix: "",
			},
			DatabasePath: "",
			Shared:       true,
		}))
		require.NotNil(t, c)
		require.NoError(t, err)

		tableName := c.GetTableName(testModelName)
		assert.Equal(t, testModelName, tableName)
	})

	// Attempt to remove a file created during the test
	t.Cleanup(func() {
		_ = os.Remove("datastore.db")
	})
}
