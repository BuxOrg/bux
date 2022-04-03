package cachestore

import (
	"context"
	"testing"
	"time"

	"github.com/mrz1836/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_WriteLock will test the method WriteLock()
func TestClient_WriteLock(t *testing.T) {

	testCases := getInMemoryTestCases(t)
	for _, testCase := range testCases {
		t.Run(testCase.name+" - missing lock key", func(t *testing.T) {
			var secret string
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WriteLock(context.Background(), "", 30)
			assert.Equal(t, "", secret)
			assert.Error(t, err)
			assert.ErrorAs(t, err, &ErrKeyRequired)
		})

		t.Run(testCase.name+" - valid lock", func(t *testing.T) {
			var secret string
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WriteLock(context.Background(), testKey, 30)
			assert.Equal(t, 64, len(secret))
			assert.NoError(t, err)

			defer func() {
				_, _ = c.ReleaseLock(context.Background(), testKey, secret)
			}()
		})

		t.Run(testCase.name+" - lock conflict", func(t *testing.T) {
			var secret string
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WriteLock(context.Background(), testKey, 30)
			assert.Equal(t, 64, len(secret))
			assert.NoError(t, err)

			defer func() {
				_, _ = c.ReleaseLock(context.Background(), testKey, secret)
			}()

			// Lock exists with different secret
			secret, err = c.WriteLock(context.Background(), testKey, 30)
			assert.Equal(t, "", secret)
			assert.ErrorAs(t, err, &ErrLockExists)
		})
	}

	// todo: add redis lock tests
}

// TestClient_ReleaseLock will test the method ReleaseLock()
func TestClient_ReleaseLock(t *testing.T) {

	testCases := getInMemoryTestCases(t)
	for _, testCase := range testCases {
		t.Run(testCase.name+" - missing lock key", func(t *testing.T) {
			var success bool
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			success, err = c.ReleaseLock(context.Background(), "", "some-value")
			assert.Equal(t, false, success)
			assert.Error(t, err)
			assert.ErrorAs(t, err, &ErrKeyRequired)
		})

		t.Run(testCase.name+" - missing secret", func(t *testing.T) {
			var success bool
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			success, err = c.ReleaseLock(context.Background(), testKey, "")
			assert.Equal(t, false, success)
			assert.Error(t, err)
			assert.ErrorAs(t, err, &ErrSecretRequired)
		})

		t.Run(testCase.name+" - valid release", func(t *testing.T) {
			var secret string
			var success bool
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WriteLock(context.Background(), testKey, 30)
			assert.Equal(t, 64, len(secret))
			assert.NoError(t, err)

			success, err = c.ReleaseLock(context.Background(), testKey, secret)
			assert.Equal(t, true, success)
			assert.NoError(t, err)
		})

		t.Run(testCase.name+" - invalid secret", func(t *testing.T) {
			var secret string
			var success bool
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WriteLock(context.Background(), testKey, 30)
			assert.Equal(t, 64, len(secret))
			assert.NoError(t, err)

			success, err = c.ReleaseLock(context.Background(), testKey, secret+"-bad-key")
			assert.Equal(t, false, success)
			assert.Error(t, err)
			assert.ErrorIs(t, err, cache.ErrLockMismatch)
		})
	}

	// todo: add redis lock tests
}

// TestClient_WaitWriteLock will test the method WaitWriteLock()
func TestClient_WaitWriteLock(t *testing.T) {

	testCases := getInMemoryTestCases(t)
	for _, testCase := range testCases {

		t.Run(testCase.name+" - missing lock key", func(t *testing.T) {
			var ctx context.Context

			var secret string
			c, err := NewClient(ctx, testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WaitWriteLock(ctx, "", 30, 10)
			assert.Equal(t, "", secret)
			assert.Error(t, err)
			assert.ErrorAs(t, err, &ErrKeyRequired)
		})

		t.Run(testCase.name+" - missing ttw", func(t *testing.T) {
			var ctx context.Context
			var secret string
			c, err := NewClient(ctx, testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WaitWriteLock(ctx, testKey, 30, 0)
			assert.Equal(t, "", secret)
			assert.Error(t, err)
			assert.ErrorAs(t, err, &ErrTTWCannotBeEmpty)
		})

		t.Run(testCase.name+" - valid lock", func(t *testing.T) {
			var ctx context.Context
			var secret string
			c, err := NewClient(ctx, testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WaitWriteLock(ctx, testKey, 30, 5)
			assert.Equal(t, 64, len(secret))
			assert.NoError(t, err)

			defer func() {
				_, _ = c.ReleaseLock(ctx, testKey, secret)
			}()
		})

		t.Run(testCase.name+" - lock jammed for a few seconds", func(t *testing.T) {
			var ctx context.Context
			var secret string
			c, err := NewClient(ctx, testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WaitWriteLock(ctx, testKey, 2, 5)
			assert.Equal(t, 64, len(secret))
			assert.NoError(t, err)

			defer func() {
				_, _ = c.ReleaseLock(ctx, testKey, secret)
			}()

			testCase.FastForward(6 * time.Second)

			secret, err = c.WaitWriteLock(ctx, testKey, 10, 5)
			assert.Equal(t, 64, len(secret))
			assert.NoError(t, err)

			defer func() {
				_, _ = c.ReleaseLock(ctx, testKey, secret)
			}()
		})

		t.Run(testCase.name+" - lock jammed, never completes", func(t *testing.T) {
			var ctx context.Context
			var secret string
			c, err := NewClient(ctx, testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			secret, err = c.WaitWriteLock(ctx, testKey, 30, 5)
			assert.Equal(t, 64, len(secret))
			assert.NoError(t, err)

			defer func() {
				_, _ = c.ReleaseLock(ctx, testKey, secret)
			}()

			secret, err = c.WaitWriteLock(ctx, testKey, 10, 2)
			assert.Equal(t, "", secret)
			assert.Error(t, err)
		})
	}
}
