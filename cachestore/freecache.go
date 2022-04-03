package cachestore

import (
	"errors"
	"runtime/debug"

	"github.com/coocood/freecache"
	"github.com/mrz1836/go-cache"
)

// loadFreeCache will load the FreeCache client
func loadFreeCache() *freecache.Cache {
	// In bytes, where 1024 * 1024 represents a single Megabyte, and 100 * 1024*1024 represents 100 Megabytes.
	cacheSize := 100 * 1024 * 1024
	c := freecache.NewCache(cacheSize)
	debug.SetGCPercent(20)
	return c
}

// writeLockFreeCache will write a lock record into memory using a secret and expiration
//
// ttl is in seconds
func writeLockFreeCache(freeCacheClient *freecache.Cache, lockKey, secret string, ttl int64) (bool, error) {

	// Test the key and secret
	if err := validateLockValues(lockKey, secret); err != nil {
		return false, err
	}

	// Try to get an existing lock (if it fails, make a new lock)
	lockKeyBytes := []byte(lockKey)
	secretBytes := []byte(secret)
	data, err := freeCacheClient.Get(lockKeyBytes)
	if err != nil && errors.Is(err, freecache.ErrNotFound) {
		return true, freeCacheClient.Set(lockKeyBytes, secretBytes, int(ttl))
	} else if err != nil {
		return false, err
	} else if err == nil && len(data) == 0 { // No lock found
		return true, freeCacheClient.Set(lockKeyBytes, secretBytes, int(ttl))
	}

	// Check secret
	if string(data) != secret { // Secret mismatch (lock exists with different secret)
		return false, cache.ErrLockMismatch
	}

	// Same secret / lock again?
	return true, freeCacheClient.Set(lockKeyBytes, secretBytes, int(ttl))
}

// releaseLockFreeCache will attempt to release a lock if it exists and matches the given secret
func releaseLockFreeCache(freeCacheClient *freecache.Cache, lockKey, secret string) (bool, error) {

	// Test the key and secret
	if err := validateLockValues(lockKey, secret); err != nil {
		return false, err
	}

	// Try to get an existing lock (if it fails, lock does not exist)
	lockKeyBytes := []byte(lockKey)
	data, err := freeCacheClient.Get(lockKeyBytes)
	if err != nil && errors.Is(err, freecache.ErrNotFound) {
		return true, nil
	} else if err != nil {
		return false, err
	} else if err == nil && len(data) == 0 { // No lock found
		return true, nil
	}

	// Check secret if found
	if string(data) == secret { // If it matches, remove the key
		freeCacheClient.Del(lockKeyBytes)
		return true, nil
	}

	// Key found does not match the secret, do not remove
	return false, cache.ErrLockMismatch
}
