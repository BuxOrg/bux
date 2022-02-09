package cachestore

import (
	"context"
	"testing"

	"github.com/BuxOrg/bux/tester"
	"github.com/stretchr/testify/require"
)

// Test_loadRedisClient will test the method loadRedisClient()
func Test_loadRedisClient(t *testing.T) {
	t.Parallel()

	t.Run("no config set", func(t *testing.T) {
		c, err := loadRedisClient(context.Background(), nil, false)
		require.Nil(t, c)
		require.Error(t, err)
	})

	t.Run("no redis url set", func(t *testing.T) {
		c, err := loadRedisClient(context.Background(), &RedisConfig{
			URL: "",
		}, false)
		require.Nil(t, c)
		require.Error(t, err)
	})

	t.Run("bad redis url set, connect will fail", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test: redis is required")
		}
		c, err := loadRedisClient(context.Background(), &RedisConfig{
			URL:            "redis://badurl:2343",
			DependencyMode: true,
		}, true)
		require.Nil(t, c)
		require.Error(t, err)
	})

	t.Run("redis url set", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test: redis is required")
		}
		c, err := loadRedisClient(context.Background(), &RedisConfig{
			URL: testLocalConnectionURL,
		}, false)
		require.NotNil(t, c)
		require.NoError(t, err)
		c.Close()
	})

	t.Run("redis url set, new relic enabled", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test: redis is required")
		}

		// Create new relic tx
		ctx := tester.GetNewRelicCtx(t, testAppName, testTxn)

		c, err := loadRedisClient(ctx, &RedisConfig{
			URL: testLocalConnectionURL,
		}, true)
		require.NotNil(t, c)
		require.NoError(t, err)
		c.Close()
	})
}
