package cachestore

import (
	"context"
	"testing"
	"time"

	"github.com/BuxOrg/bux/tester"
	"github.com/BuxOrg/bux/utils"
	"github.com/dgraph-io/ristretto"
	"github.com/mrz1836/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRistrettoClient makes a new testing client
func newTestRistrettoClient(t *testing.T) ClientInterface {
	c, err := NewClient(context.Background(), WithRistretto(DefaultRistrettoConfig()))
	require.NotNil(t, c)
	require.NoError(t, err)
	return c
}

// Test_loadRistrettoClient will test the method loadRistrettoClient()
func Test_loadRistrettoClient(t *testing.T) {
	t.Parallel()

	t.Run("no config set", func(t *testing.T) {
		c, err := loadRistrettoClient(context.Background(), nil, false)
		require.Nil(t, c)
		require.Error(t, err)
	})

	t.Run("bad config values", func(t *testing.T) {
		c, err := loadRistrettoClient(context.Background(), &ristretto.Config{}, true)
		require.Nil(t, c)
		require.Error(t, err)
	})

	t.Run("valid config, new relic enabled", func(t *testing.T) {
		ctx := tester.GetNewRelicCtx(t, testAppName, testTxn)
		c, err := loadRistrettoClient(ctx, DefaultRistrettoConfig(), true)
		require.NotNil(t, c)
		require.NoError(t, err)
	})
}

// Test_writeLockRistretto will test the method writeLockRistretto()
func Test_writeLockRistretto(t *testing.T) {
	t.Parallel()

	t.Run("missing key", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestRistrettoClient(t)

		var success bool
		success, err = writeLockRistretto(c.Ristretto(), "", testSecret, 1, 30)
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("missing secret", func(t *testing.T) {
		c := newTestRistrettoClient(t)

		success, err := writeLockRistretto(c.Ristretto(), testKey, "", 2, 30)
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSecretRequired)
	})

	t.Run("valid lock", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestRistrettoClient(t)

		var success bool
		success, err = writeLockRistretto(c.Ristretto(), testKey, testSecret, 1, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		defer func() {
			_, _ = releaseLockRistretto(c.Ristretto(), testKey, testSecret)
		}()
	})

	t.Run("create lock, try creating second lock", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestRistrettoClient(t)

		// Create a new lock
		var success bool
		success, err = writeLockRistretto(c.Ristretto(), testKey, testSecret, 1, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		// Release at the end of this function
		defer func() {
			_, err = releaseLockRistretto(c.Ristretto(), testKey, testSecret)
			assert.NoError(t, err)
		}()

		// Sleep!
		time.Sleep(1 * time.Second)

		// Try creating lock for same key, different secret
		success, err = writeLockRistretto(c.Ristretto(), testKey, testSecret+"different", 1, 30)
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrLockMismatch)

		// Sleep!
		time.Sleep(1 * time.Second)

		// Try creating a valid lock again (same secret, extending the TTL)
		success, err = writeLockRistretto(c.Ristretto(), testKey, testSecret, 1, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)
	})
}

// Test_releaseLockRistretto will test the method releaseLockRistretto()
func Test_releaseLockRistretto(t *testing.T) {
	t.Parallel()

	t.Run("missing key", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestRistrettoClient(t)

		var success bool
		success, err = releaseLockRistretto(c.Ristretto(), "", testSecret)
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("missing secret", func(t *testing.T) {
		c := newTestRistrettoClient(t)
		success, err := releaseLockRistretto(c.Ristretto(), testKey, "")
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSecretRequired)
	})

	t.Run("release valid lock", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestRistrettoClient(t)

		var success bool
		success, err = writeLockRistretto(c.Ristretto(), testKey, testSecret, 1, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		// remove the lock (found)
		success, err = releaseLockRistretto(c.Ristretto(), testKey, testSecret)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		// No lock found, still ok if fired a second time
		success, err = releaseLockRistretto(c.Ristretto(), testKey, testSecret)
		assert.Equal(t, true, success)
		assert.NoError(t, err)
	})

	t.Run("lock mismatch", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestRistrettoClient(t)

		var success bool
		success, err = writeLockRistretto(c.Ristretto(), testKey, testSecret, 1, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		// test secret
		success, err = releaseLockRistretto(c.Ristretto(), testKey, testSecret+"wrong")
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrLockMismatch)
	})
}

// TestDefaultRistrettoConfig will test the method DefaultRistrettoConfig()
func TestDefaultRistrettoConfig(t *testing.T) {
	t.Parallel()

	t.Run("default configuration", func(t *testing.T) {
		c := DefaultRistrettoConfig()
		require.NotNil(t, c)
		assert.Equal(t, int64(64), c.BufferItems)
		assert.Equal(t, int64(10000000), c.NumCounters)
		assert.Equal(t, int64(1073741824), c.MaxCost)
		assert.Equal(t, false, c.IgnoreInternalCost)
		assert.Equal(t, false, c.Metrics)
	})
}
