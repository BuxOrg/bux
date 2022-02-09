package cachestore

import (
	"context"
	"testing"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestMcacheClient makes a new testing client
func newTestMcacheClient(t *testing.T) ClientInterface {
	c, err := NewClient(context.Background(), WithMcache())
	require.NotNil(t, c)
	require.NoError(t, err)
	return c
}

// Test_writeLockMcache will test the method writeLockMcache()
func Test_writeLockMcache(t *testing.T) {

	t.Run("missing key", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestMcacheClient(t)

		var success bool
		success, err = writeLockMcache(c.MCache(), "", testSecret, 30)
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("missing secret", func(t *testing.T) {
		c := newTestMcacheClient(t)

		success, err := writeLockMcache(c.MCache(), testKey, "", 30)
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSecretRequired)
	})

	t.Run("valid lock", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestMcacheClient(t)

		var success bool
		success, err = writeLockMcache(c.MCache(), testKey, testSecret, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		defer func() {
			_, _ = releaseLockMcache(c.MCache(), testKey, testSecret)
		}()
	})

	t.Run("create lock, try creating second lock", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestMcacheClient(t)

		// Create a new lock
		var success bool
		success, err = writeLockMcache(c.MCache(), testKey, testSecret, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		// Release at the end of this function
		defer func() {
			_, err = releaseLockMcache(c.MCache(), testKey, testSecret)
			assert.NoError(t, err)
		}()

		// Sleep!
		time.Sleep(1 * time.Second)

		// Try creating lock for same key, different secret
		success, err = writeLockMcache(c.MCache(), testKey, testSecret+"different", 30)
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrLockMismatch)

		// Sleep!
		time.Sleep(1 * time.Second)

		// Try creating a valid lock again (same secret, extending the TTL)
		success, err = writeLockMcache(c.MCache(), testKey, testSecret, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)
	})
}

// Test_releaseLockMcache will test the method releaseLockMcache()
func Test_releaseLockMcache(t *testing.T) {

	t.Run("missing key", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestMcacheClient(t)

		var success bool
		success, err = releaseLockMcache(c.MCache(), "", testSecret)
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("missing secret", func(t *testing.T) {
		c := newTestMcacheClient(t)
		success, err := releaseLockMcache(c.MCache(), testKey, "")
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSecretRequired)
	})

	t.Run("release valid lock", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestMcacheClient(t)

		var success bool
		success, err = writeLockMcache(c.MCache(), testKey, testSecret, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		// remove the lock (found)
		success, err = releaseLockMcache(c.MCache(), testKey, testSecret)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		// No lock found, still ok if fired a second time
		success, err = releaseLockMcache(c.MCache(), testKey, testSecret)
		assert.Equal(t, true, success)
		assert.NoError(t, err)
	})

	t.Run("lock mismatch", func(t *testing.T) {
		testSecret, err := utils.RandomHex(32)
		require.NoError(t, err)

		c := newTestMcacheClient(t)

		var success bool
		success, err = writeLockMcache(c.MCache(), testKey, testSecret, 30)
		assert.Equal(t, true, success)
		assert.NoError(t, err)

		// test secret
		success, err = releaseLockMcache(c.MCache(), testKey, testSecret+"wrong")
		assert.Equal(t, false, success)
		assert.Error(t, err)
		assert.ErrorIs(t, err, cache.ErrLockMismatch)
	})
}

// Test_validateLockValues will test the method validateLockValues()
func Test_validateLockValues(t *testing.T) {
	t.Parallel()

	t.Run("missing key", func(t *testing.T) {
		err := validateLockValues("", testValue)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("missing secret", func(t *testing.T) {
		err := validateLockValues(testKey, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSecretRequired)
	})

	t.Run("valid values", func(t *testing.T) {
		err := validateLockValues(testKey, testValue)
		assert.NoError(t, err)
	})
}
