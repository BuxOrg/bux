package cachestore

import (
	"errors"
	"runtime/debug"

	"github.com/coocood/freecache"
	"github.com/mrz1836/go-cache"
)

const (

	// DefaultCacheSize in bytes, where 1024 * 1024 represents a single Megabyte, and 100 * 1024*1024 represents 100 Megabytes.
	DefaultCacheSize = 100 * 1024 * 1024

	// DefaultGCPercent is the percentage when full it will run GC
	DefaultGCPercent = 20
)

// loadFreeCache will load the FreeCache client
//
// This is a default cache solution for running a local single server.
func loadFreeCache(cacheSize, percent int) (c *freecache.Cache) {

	// Set the defaults for cache size
	if cacheSize <= 0 {
		cacheSize = DefaultCacheSize
	}
	c = freecache.NewCache(cacheSize)

	// Set the default GC percent
	if percent <= 0 {
		percent = DefaultGCPercent
	}
	debug.SetGCPercent(percent)
	return
}

// writeLockFreeCache will write a lock record into memory using a secret and expiration
//
// ttl is in seconds
func writeLockFreeCache(freeCacheClient *freecache.Cache, lockKey, secret string, ttl int64) (bool, error) {

	// Try to get an existing lock (if it fails, make a new lock)
	lockKeyBytes := []byte(lockKey)
	secretBytes := []byte(secret)
	data, err := freeCacheClient.Get(lockKeyBytes)
	if err != nil && errors.Is(err, freecache.ErrNotFound) {
		return true, freeCacheClient.Set(lockKeyBytes, secretBytes, int(ttl))
	} else if err != nil {
		return false, err
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

	// Try to get an existing lock (if it fails, lock does not exist)
	lockKeyBytes := []byte(lockKey)
	data, err := freeCacheClient.Get(lockKeyBytes)
	if err != nil && errors.Is(err, freecache.ErrNotFound) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	// Check secret if found
	if string(data) == secret { // If it matches, remove the key
		freeCacheClient.Del(lockKeyBytes)
		return true, nil
	}

	// Key found does not match the secret, do not remove
	return false, cache.ErrLockMismatch
}
