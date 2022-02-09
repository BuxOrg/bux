package tester

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// todo: finish unit tests!

// TestAnyTime_Match will test the method Match()
func TestAnyTime_Match(t *testing.T) {
	// finish test
}

// TestAnyGUID_Match will test the method Match()
func TestAnyGUID_Match(t *testing.T) {
	// finish test
}

/*
// TestCreatePostgresServer will test the method CreatePostgresServer()
func TestCreatePostgresServer(t *testing.T) {
	// t.Parallel() (disabled for now)

	t.Run("valid server", func(t *testing.T) {
		server, err := CreatePostgresServer(
			testDatabasePort1,
		)
		require.NoError(t, err)
		require.NotNil(t, server)
		err = server.Stop()
		require.NoError(t, err)
	})
}*/

// TestCreateMongoServer will test the method CreateMongoServer()
func TestCreateMongoServer(t *testing.T) {
	t.Parallel()

	t.Run("valid server", func(t *testing.T) {
		server, err := CreateMongoServer(
			testMongoVersion,
		)
		require.NoError(t, err)
		require.NotNil(t, server)
		server.Stop()
	})
}

// TestCreateMySQL will test the method CreateMySQL()
func TestCreateMySQL(t *testing.T) {
	t.Parallel()

	t.Run("valid server", func(t *testing.T) {
		server, err := CreateMySQL(
			testDatabaseHost, testDatabaseName, testDatabaseUser,
			testDatabasePassword, testDatabasePort2,
		)
		require.NoError(t, err)
		require.NotNil(t, server)
		err = server.Close()
		require.NoError(t, err)
	})
}

// TestCreateMySQLTestDatabase will test the method CreateMySQLTestDatabase()
func TestCreateMySQLTestDatabase(t *testing.T) {
	t.Parallel()

	t.Run("valid db", func(t *testing.T) {
		db := CreateMySQLTestDatabase(testDatabaseName)
		require.NotNil(t, db)
		assert.Equal(t, testDatabaseName, db.Name())
	})
}

// TestSQLiteTestConfig will test the method SQLiteTestConfig()
func TestSQLiteTestConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		config := SQLiteTestConfig(t, true, true)
		require.NotNil(t, config)

		assert.Equal(t, true, config.Debug)
		assert.Equal(t, true, config.Shared)
		assert.Equal(t, 1, config.MaxIdleConnections)
		assert.Equal(t, 1, config.MaxOpenConnections)
		assert.NotEmpty(t, config.TablePrefix)
		assert.Empty(t, config.DatabasePath)
	})

	t.Run("no debug or sharing", func(t *testing.T) {
		config := SQLiteTestConfig(t, false, false)
		require.NotNil(t, config)

		assert.Equal(t, false, config.Debug)
		assert.Equal(t, false, config.Shared)
	})
}
