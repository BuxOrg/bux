package cachestore

import (
	"context"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-cache"
	"github.com/pkg/errors"
)

// WriteLock will create a unique lock/secret with a TTL (seconds) to expire
// The lockKey is unique and should be deterministic
// The secret will be automatically generated and stored in the locked key (returned)
func (c *Client) WriteLock(ctx context.Context, lockKey string, ttl int64) (string, error) {

	var secret string
	var locked bool
	var err error

	// Create a secret
	if secret, err = utils.RandomHex(32); err != nil {
		// This will "ALMOST NEVER" error out
		return "", errors.Wrap(ErrSecretGenerationFailed, err.Error())
	}

	// Lock using Redis
	if c.Engine() == Redis {
		if len(lockKey) == 0 { // This happens in mCache already
			return "", ErrKeyRequired
		}
		if locked, err = cache.WriteLock(
			ctx, c.options.redis, lockKey, secret, ttl,
		); err != nil {
			return "", errors.Wrap(ErrLockCreateFailed, err.Error())
		} else if !locked {
			return "", ErrLockExists
		}
	} else if c.Engine() == MCache { // Lock using MCache
		if locked, err = writeLockMcache(
			c.options.mCache, lockKey, secret, ttl,
		); err != nil {
			return "", errors.Wrap(ErrLockCreateFailed, err.Error())
		} else if !locked {
			return "", ErrLockExists
		}
	} else if c.Engine() == Ristretto { // Lock using Ristretto
		if locked, err = writeLockRistretto(
			c.options.ristretto, lockKey, secret, baseCostPerKey, ttl,
		); err != nil {
			return "", errors.Wrap(ErrLockCreateFailed, err.Error())
		} else if !locked {
			return "", ErrLockExists
		}
	} else { // Engine is not supported
		return "", ErrEngineNotSupported
	}

	return secret, nil
}

// WaitWriteLock will aggressively try to make a lock until the TTW (in seconds) is reached
func (c *Client) WaitWriteLock(ctx context.Context, lockKey string, ttl, ttw int64) (string, error) {

	var secret string

	// Test the values
	if len(lockKey) == 0 {
		return secret, ErrKeyRequired
	} else if ttw <= 0 {
		return secret, ErrTTWCannotBeEmpty
	}

	// Create the end time for the loop
	end := time.Now().Add(time.Duration(ttw) * time.Second)

	// Loop until we have a secret, or we are passed the end time
	for {
		if secret, _ = c.WriteLock(
			ctx, lockKey, ttl,
		); len(secret) > 0 || time.Now().After(end) {
			break
		} else {
			time.Sleep(lockRetrySleepTime)
		}
	}

	// No secret, lock creating failed or did not complete
	if len(secret) == 0 {
		return "", ErrLockCreateFailed
	}

	return secret, nil
}

// ReleaseLock will release a given lock key only if the secret matches
func (c *Client) ReleaseLock(ctx context.Context, lockKey, secret string) (bool, error) {

	// Release the lock
	if c.Engine() == Redis {
		return cache.ReleaseLock(ctx, c.options.redis, lockKey, secret)
	} else if c.Engine() == MCache {
		return releaseLockMcache(c.options.mCache, lockKey, secret)
	} else if c.Engine() == Ristretto {
		return releaseLockRistretto(c.options.ristretto, lockKey, secret)
	}

	// Engine is not supported
	return false, ErrEngineNotSupported
}
