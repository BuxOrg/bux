package cachestore

import (
	"context"
	"testing"
	"time"

	"github.com/BuxOrg/bux/tester"
	"github.com/OrlovEvgeny/go-mcache"
	"github.com/dgraph-io/ristretto"
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
		assert.Nil(t, defaults.mCache)
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
		opts := []ClientOps{WithNewRelic(), WithMcache()}
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
		opts := []ClientOps{WithDebugging(), WithMcache()}
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
		assert.Nil(t, c)
		assert.Error(t, err)
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

// TestWithMcache will test the method WithMcache()
func TestWithMcache(t *testing.T) {
	t.Run("get opts", func(t *testing.T) {
		opt := WithMcache()
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply basic config", func(t *testing.T) {

		opts := []ClientOps{WithDebugging(), WithMcache()}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, MCache, c.Engine())
		assert.IsType(t, &mcache.CacheDriver{}, c.MCache())
	})
}

// TestWithRistretto will test the method WithRistretto()
func TestWithRistretto(t *testing.T) {
	t.Run("get opts", func(t *testing.T) {
		opt := WithRistretto(&ristretto.Config{
			NumCounters: 100,
			MaxCost:     100,
			BufferItems: 1,
		})
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply basic config", func(t *testing.T) {

		opts := []ClientOps{WithDebugging(), WithRistretto(&ristretto.Config{
			NumCounters: 100,
			MaxCost:     100,
			BufferItems: 1,
		})}
		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, Ristretto, c.Engine())
		assert.IsType(t, &ristretto.Cache{}, c.Ristretto())
	})
}

// TestWithMcacheConnection will test the method WithMcacheConnection()
func TestWithMcacheConnection(t *testing.T) {
	t.Run("get opts", func(t *testing.T) {
		opt := WithMcacheConnection(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply empty connection", func(t *testing.T) {
		opts := []ClientOps{WithDebugging(), WithMcacheConnection(nil)}
		c, err := NewClient(context.Background(), opts...)
		assert.Nil(t, c)
		assert.Error(t, err)
	})

	t.Run("apply existing external connection", func(t *testing.T) {
		newClient := mcache.New()
		require.NotNil(t, newClient)

		opts := []ClientOps{WithDebugging(), WithMcacheConnection(newClient)}

		c, err := NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, MCache, c.Engine())
		assert.IsType(t, &mcache.CacheDriver{}, c.MCache())
	})
}

// TestWithRistrettoConnection will test the method WithRistrettoConnection()
func TestWithRistrettoConnection(t *testing.T) {
	t.Run("get opts", func(t *testing.T) {
		opt := WithRistretto(nil)
		assert.IsType(t, *new(ClientOps), opt)
	})

	t.Run("apply empty connection", func(t *testing.T) {
		opts := []ClientOps{WithDebugging(), WithRistretto(nil)}
		c, err := NewClient(context.Background(), opts...)
		assert.Nil(t, c)
		assert.Error(t, err)
	})

	t.Run("apply existing external connection", func(t *testing.T) {
		newClient, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: 100,
			MaxCost:     1,
			BufferItems: 10,
		})
		require.NotNil(t, newClient)
		require.NoError(t, err)

		opts := []ClientOps{WithDebugging(), WithRistrettoConnection(newClient)}

		var c ClientInterface
		c, err = NewClient(context.Background(), opts...)
		require.NotNil(t, c)
		require.NoError(t, err)

		assert.Equal(t, Ristretto, c.Engine())
		assert.IsType(t, &ristretto.Cache{}, c.Ristretto())
		assert.Nil(t, c.RistrettoConfig())
	})
}
