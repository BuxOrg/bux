package cachestore

import (
	"context"
	"testing"
	"time"

	"github.com/BuxOrg/bux/logger"
	"github.com/BuxOrg/bux/tester"
	"github.com/coocood/freecache"
	"github.com/mrz1836/go-cache"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultClientOptions will test the method defaultClientOptions()
func TestDefaultClientOptions(t *testing.T) {
	t.Run("ensure default values", func(t *testing.T) {
		defaults := defaultClientOptions()
		require.NotNil(t, defaults)
		assert.Equal(t, Empty, defaults.engine)
		assert.Equal(t, false, defaults.newRelicEnabled)
		assert.Nil(t, defaults.redis)
		require.NotNil(t, defaults.redisConfig)
	})
}

// TestClientOptions_GetTxnCtx will test the method getTxnCtx()
func TestClientOptions_GetTxnCtx(t *testing.T) {
	t.Run("no txn found", func(t *testing.T) {
		defaults := defaultClientOptions()
		require.NotNil(t, defaults)
		defaults.newRelicEnabled = true

		ctx := defaults.getTxnCtx(context.Background())
		require.NotNil(t, ctx)

		txn := newrelic.FromContext(ctx)
		assert.Nil(t, txn)
	})

	t.Run("txn found", func(t *testing.T) {
		ctx := tester.GetNewRelicCtx(t, "test", "test-tx")
		txn := newrelic.FromContext(ctx)
		assert.NotNil(t, txn)
	})
}

// TestWithNewRelic will test the method WithNewRelic()
func TestWithNewRelic(t *testing.T) {

	t.Run("get opts", func(t *testing.T) {
		opt := WithNewRelic()
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply opts", func(t *testing.T) {
		opts := []ClientOps{WithNewRelic(), WithFreeCache()}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, true, c.IsNewRelicEnabled())
	})
}

// TestWithDebugging will test the method WithDebugging()
func TestWithDebugging(t *testing.T) {

	t.Run("get opts", func(t *testing.T) {
		opt := WithDebugging()
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply opts", func(t *testing.T) {
		opts := []ClientOps{WithDebugging(), WithFreeCache()}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, true, c.IsDebug())
	})
}

// TestWithRedis will test the method WithRedis()
func TestWithRedis(t *testing.T) {

	t.Run("get opts", func(t *testing.T) {
		opt := WithRedis(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply empty config, error", func(t *testing.T) {
		config := &RedisConfig{}
		opts := []ClientOps{WithDebugging(), WithRedis(config)}
		c, err := NewClient(context.Background(), opts...)
		assert.Nil(t, c)
		assert.Error(t, err)
	})

	t.Run("apply basic local config", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		config := &RedisConfig{
			URL: testLocalConnectionURL,
		}
		opts := []ClientOps{WithDebugging(), WithRedis(config)}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, DefaultRedisMaxIdleTimeout.String(), c.RedisConfig().MaxIdleTimeout.String())
		assert.Equal(t, Redis, c.Engine())
	})

	t.Run("missing redis prefix", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		config := &RedisConfig{
			URL: "localhost:6379",
		}
		opts := []ClientOps{WithDebugging(), WithRedis(config)}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, DefaultRedisMaxIdleTimeout.String(), c.RedisConfig().MaxIdleTimeout.String())
		assert.Equal(t, Redis, c.Engine())
	})

	t.Run("apply verbose config", func(t *testing.T) {

		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		config := &RedisConfig{
			URL:                   testLocalConnectionURL,
			MaxActiveConnections:  10,
			MaxConnectionLifetime: 10 * time.Second,
			MaxIdleConnections:    10,
			MaxIdleTimeout:        10 * time.Second,
			DependencyMode:        true,
			UseTLS:                false,
		}
		opts := []ClientOps{WithDebugging(), WithRedis(config)}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, Redis, c.Engine())
	})
}

// TestWithRedisConnection will test the method WithRedisConnection()
func TestWithRedisConnection(t *testing.T) {
	t.Run("get opts", func(t *testing.T) {
		opt := WithRedisConnection(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply empty redis connection", func(t *testing.T) {
		opts := []ClientOps{WithDebugging(), WithRedisConnection(nil)}
		c, err := NewClient(context.Background(), opts...)
		assert.NotNil(t, c)
		assert.NoError(t, err)
		assert.Equal(t, FreeCache, c.Engine())
	})

	t.Run("apply existing external redis connection", func(t *testing.T) {
		newClient, err := loadRedisClient(context.Background(), &RedisConfig{
			URL: testLocalConnectionURL,
		}, false)
		require.NoError(t, err)

		opts := []ClientOps{WithDebugging(), WithRedisConnection(newClient)}

		var c ClientInterface
		c, err = NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, Redis, c.Engine())
		assert.IsType(t, &cache.Client{}, c.Redis())
		assert.Nil(t, c.RedisConfig())
	})
}

// TestWithFreeCache will test the method WithFreeCache()
func TestWithFreeCache(t *testing.T) {
	t.Run("get opts", func(t *testing.T) {
		opt := WithFreeCache()
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("use the default configuration", func(t *testing.T) {

		opts := []ClientOps{WithDebugging(), WithFreeCache()}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, FreeCache, c.Engine())
		assert.IsType(t, &freecache.Cache{}, c.FreeCache())
	})
}

// TestWithFreeCacheConnection will test the method WithFreeCacheConnection()
func TestWithFreeCacheConnection(t *testing.T) {
	t.Run("get opts", func(t *testing.T) {
		opt := WithFreeCacheConnection(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("use an existing connection", func(t *testing.T) {

		freeClient := loadFreeCache(DefaultCacheSize, DefaultGCPercent)

		opts := []ClientOps{WithDebugging(), WithFreeCacheConnection(freeClient)}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, FreeCache, c.Engine())
		assert.Equal(t, freeClient, c.FreeCache())
	})
}

// TestWithLogger will test the method WithLogger()
func TestWithLogger(t *testing.T) {
	t.Parallel()

	t.Run("check type", func(t *testing.T) {
		opt := WithLogger(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("test applying nil", func(t *testing.T) {
		options := &clientOptions{}
		opt := WithLogger(nil)
		opt(options)
		assert.Nil(t, options.logger)
	})

	t.Run("test applying option", func(t *testing.T) {
		options := &clientOptions{}
		customClient := logger.NewLogger(true, 4)
		opt := WithLogger(customClient)
		opt(options)
		assert.Equal(t, customClient, options.logger)
	})
}
