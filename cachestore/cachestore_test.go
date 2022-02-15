package cachestore

import (
	"context"
	"testing"
	"time"

	"github.com/BuxOrg/bux/tester"
	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-cache"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testIdleTimeout          = 240 * time.Second
	testKey                  = "test-key"
	testLocalConnectionURL   = RedisPrefix + "localhost:6379"
	testMaxActiveConnections = 0
	testMaxConnLifetime      = 60 * time.Second
	testMaxIdleConnections   = 10
	testValue                = "test-value"
	testAppName              = "test-app"
	testTxn                  = "test-txn"
)

// genericStruct is an example struct for testing
type genericStruct struct {
	StringField string
	IntField    int
	BoolField   bool
	FloatField  float64
}

// newMockRedisClient will create a new redis mock client
func newMockRedisClient(t *testing.T) (ClientInterface, *redigomock.Conn) {
	redisClient, conn := tester.LoadMockRedis(
		testIdleTimeout, testMaxConnLifetime, testMaxActiveConnections, testMaxIdleConnections,
	)
	require.NotNil(t, redisClient)
	require.NotNil(t, conn)

	c, err := NewClient(context.Background(), WithRedisConnection(redisClient))
	require.NotNil(t, c)
	require.NoError(t, err)
	return c, conn
}

// cacheTestCase is the test case struct
type cacheTestCase struct {
	name   string
	engine Engine
	opts   ClientOps
}

// Test cases for all in-memory cachestore engines
var cacheTestCases = []cacheTestCase{
	{name: "[mcache] [in-memory]", engine: MCache, opts: WithMcache()},
	{name: "[ristretto] [in-memory]", engine: Ristretto, opts: WithRistretto(DefaultRistrettoConfig())},
}

func TestCachestore_Interface(t *testing.T) {
	t.Run("mocked - valid datastore config", func(t *testing.T) {
		ctx := context.Background()
		client, conn := newMockRedisClient(t)

		require.NotNil(t, client)
		require.NotNil(t, conn)

		fees := "512"

		// Set command
		setCmd := conn.Command(cache.SetCommand, testKey, fees).Expect(fees)
		err := client.Set(ctx, testKey, fees)
		require.NoError(t, err)
		assert.Equal(t, true, setCmd.Called)

		// Get command
		getCmd := conn.Command(cache.GetCommand, testKey).Expect(fees)
		getFees, err2 := client.Get(ctx, testKey)
		require.NoError(t, err2)
		assert.Equal(t, true, getCmd.Called)

		assert.Equal(t, fees, getFees)
	})

	t.Run("valid datastore config", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test: redis is required")
		}

		c, err := NewClient(context.Background(),
			WithRedis(&RedisConfig{
				URL: testLocalConnectionURL,
			}),
			WithDebugging(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)

		ctx := context.Background()

		fees := "512"

		// Set command
		err = c.Set(ctx, testKey, fees)
		require.NoError(t, err)

		// Get command
		getFees, err2 := c.Get(ctx, testKey)
		require.NoError(t, err2)

		assert.Equal(t, fees, getFees)
	})
}

func TestCachestore_Models(t *testing.T) {
	t.Run("valid datastore config", func(t *testing.T) {
		ctx := context.Background()
		if testing.Short() {
			t.Skip("skipping test: redis is required")
		}

		c, err := NewClient(context.Background(),
			WithRedis(&RedisConfig{
				URL: testLocalConnectionURL,
			}),
			WithDebugging(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)

		fee := &utils.FeeUnit{
			Satoshis: 500,
			Bytes:    1000,
		}

		err = c.SetModel(ctx, testKey, fee, 1*time.Minute)
		require.NoError(t, err)

		getFee := new(utils.FeeUnit)
		err = c.GetModel(ctx, testKey, getFee)
		require.NoError(t, err)
		t.Log(getFee)
		assert.Equal(t, 500, getFee.Satoshis)
		assert.Equal(t, 1000, getFee.Bytes)
	})
}
