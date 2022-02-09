package cachestore

import (
	"time"

	"github.com/OrlovEvgeny/go-mcache"
	"github.com/mrz1836/go-cache"
)

// writeLockMcache will write a lock record into memory using a secret and expiration
//
// ttl is in seconds
func writeLockMcache(mCache *mcache.CacheDriver, lockKey, secret string, ttl int64) (bool, error) {

	// Test the key and secret
	if err := validateLockValues(lockKey, secret); err != nil {
		return false, err
	}

	// Try to get an existing lock (if it fails, make a new lock)
	data, ok := mCache.Get(lockKey)
	if !ok { // No lock found
		return mCacheSet(mCache, lockKey, secret, ttl)
	}

	// Check secret
	if data.(string) != secret { // Secret mismatch (lock exists with different secret)
		return false, cache.ErrLockMismatch
	}

	// Same secret / lock again?
	return mCacheSet(mCache, lockKey, secret, ttl)
}

// releaseLockMcache will attempt to release a lock if it exists and matches the given secret
func releaseLockMcache(mCache *mcache.CacheDriver, lockKey, secret string) (bool, error) {

	// Test the key and secret
	if err := validateLockValues(lockKey, secret); err != nil {
		return false, err
	}

	// Try to get an existing lock (if it fails, lock does not exist)
	data, ok := mCache.Get(lockKey)
	if !ok { // No lock found
		return true, nil
	}

	// Check secret if found
	if data.(string) == secret { // If it matches, remove the key
		mCache.Remove(lockKey)
		return true, nil
	}

	// Key found does not match the secret, do not remove
	return false, cache.ErrLockMismatch
}

// mCacheSet will set a key in mCache and check for the error
//
// ttl is in seconds
func mCacheSet(mCache *mcache.CacheDriver, key string, value interface{}, ttl int64) (bool, error) {
	if err := mCache.Set(
		key, value, time.Second*time.Duration(ttl),
	); err != nil {
		return false, err
	}
	return true, nil
}

// validateLockValues will validate and test the lock/secret values
func validateLockValues(lockKey, secret string) error {

	// Require a key to be present
	if len(lockKey) == 0 {
		return ErrKeyRequired
	}

	// Require a secret to be present
	if len(secret) == 0 {
		return ErrSecretRequired
	}
	return nil
}
