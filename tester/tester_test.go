package tester

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	// testDatabasePort1    = 23902
	testDatabaseHost     = "localhost"
	testDatabaseName     = "test"
	testDatabasePassword = "tester-pw"
	testDatabasePort2    = 23903
	testDatabaseUser     = "tester"
	testDomain           = "domain.com"
	testMongoVersion     = "4.2.1"
	testRedisConnection  = "redis://localhost:6379"
)

// TestRandomTablePrefix will test the method RandomTablePrefix()
func TestRandomTablePrefix(t *testing.T) {
	t.Parallel()

	t.Run("valid prefix", func(t *testing.T) {
		prefix := RandomTablePrefix(t)
		assert.Equal(t, 17, len(prefix))
	})
}
