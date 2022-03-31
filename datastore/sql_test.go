package datastore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// todo: finish unit tests!

// TestClient_openSQLDatabase will test the method openSQLDatabase()
func TestClient_openSQLDatabase(t *testing.T) {
	// finish test
}

// TestClient_openSQLiteDatabase will test the method openSQLiteDatabase()
func TestClient_openSQLiteDatabase(t *testing.T) {
	// finish test
}

// TestClient_getDNS will test the method getDNS()
func TestClient_getDNS(t *testing.T) {
	// finish test
}

// TestClient_getDialector will test the method getDialector()
func TestClient_getDialector(t *testing.T) {
	// finish test
}

// TestClient_mySQLDialector will test the method mySQLDialector()
func TestClient_mySQLDialector(t *testing.T) {
	// finish test
}

// TestClient_postgreSQLDialector will test the method postgreSQLDialector()
func TestClient_postgreSQLDialector(t *testing.T) {
	// finish test
}

// TestClient_getSourceDatabase will test the method getSourceDatabase()
func TestClient_getSourceDatabase(t *testing.T) {

	t.Run("single write db", func(t *testing.T) {
		source, configs := getSourceDatabase(
			[]*SQLConfig{
				{
					CommonConfig: CommonConfig{
						Debug:                 true,
						MaxConnectionIdleTime: 10 * time.Second,
						MaxConnectionTime:     10 * time.Second,
						MaxIdleConnections:    1,
						MaxOpenConnections:    1,
						TablePrefix:           "test",
					},
					Driver:    MySQL.String(),
					Host:      "host-write.domain.com",
					Name:      "db_name",
					Password:  "test",
					Port:      "3306",
					Replica:   false,
					TimeZone:  "UTC",
					TxTimeout: 10 * time.Second,
					User:      "test",
				},
			},
		)
		require.NotNil(t, source)
		require.Equal(t, 0, len(configs))
		assert.Equal(t, false, source.Replica)
		assert.Equal(t, "host-write.domain.com", source.Host)
	})

	t.Run("read vs write", func(t *testing.T) {
		source, configs := getSourceDatabase(
			[]*SQLConfig{
				{
					CommonConfig: CommonConfig{
						Debug:                 true,
						MaxConnectionIdleTime: 10 * time.Second,
						MaxConnectionTime:     10 * time.Second,
						MaxIdleConnections:    1,
						MaxOpenConnections:    1,
						TablePrefix:           "test",
					},
					Driver:    MySQL.String(),
					Host:      "host-write.domain.com",
					Name:      "db_name",
					Password:  "test",
					Port:      "3306",
					Replica:   false,
					TimeZone:  "UTC",
					TxTimeout: 10 * time.Second,
					User:      "test",
				},
				{
					CommonConfig: CommonConfig{
						Debug:                 true,
						MaxConnectionIdleTime: 10 * time.Second,
						MaxConnectionTime:     10 * time.Second,
						MaxIdleConnections:    1,
						MaxOpenConnections:    1,
						TablePrefix:           "test",
					},
					Driver:    MySQL.String(),
					Host:      "host-read.domain.com",
					Name:      "db_name",
					Password:  "test",
					Port:      "3306",
					Replica:   true,
					TimeZone:  "UTC",
					TxTimeout: 10 * time.Second,
					User:      "test",
				},
			},
		)
		require.NotNil(t, source)

		assert.Equal(t, false, source.Replica)
		assert.Equal(t, "host-write.domain.com", source.Host)

		assert.Equal(t, 1, len(configs))
		assert.Equal(t, true, configs[0].Replica)
		assert.Equal(t, "host-read.domain.com", configs[0].Host)
	})

	t.Run("only replica, no source found", func(t *testing.T) {
		source, configs := getSourceDatabase(
			[]*SQLConfig{
				{
					CommonConfig: CommonConfig{
						Debug:                 true,
						MaxConnectionIdleTime: 10 * time.Second,
						MaxConnectionTime:     10 * time.Second,
						MaxIdleConnections:    1,
						MaxOpenConnections:    1,
						TablePrefix:           "test",
					},
					Driver:    MySQL.String(),
					Host:      "host-read.domain.com",
					Name:      "db_name",
					Password:  "test",
					Port:      "3306",
					Replica:   true,
					TimeZone:  "UTC",
					TxTimeout: 10 * time.Second,
					User:      "test",
				},
			},
		)
		require.Nil(t, source)
		assert.Equal(t, 1, len(configs))
		assert.Equal(t, true, configs[0].Replica)
		assert.Equal(t, "host-read.domain.com", configs[0].Host)
	})
}

// TestClient_getGormSessionConfig will test the method getGormSessionConfig()
func TestClient_getGormSessionConfig(t *testing.T) {
	// finish test
}

// TestClient_getGormConfig will test the method getGormConfig()
func TestClient_getGormConfig(t *testing.T) {
	// finish test
}

// TestClient_closeSQLDatabase will test the method closeSQLDatabase()
func TestClient_closeSQLDatabase(t *testing.T) {
	// finish test
}

// TestSQLConfig_sqlDefaults will test the method sqlDefaults()
func TestSQLConfig_sqlDefaults(t *testing.T) {
	// finish test
}
