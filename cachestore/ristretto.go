package cachestore

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/mrz1836/go-cache"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// DefaultRistrettoConfig will return a default configuration that can be modified
func DefaultRistrettoConfig() *ristretto.Config {
	return &ristretto.Config{
		BufferItems:        64,      // Number of keys per Get buffer
		IgnoreInternalCost: false,   // Ignore cost values, memory will grow
		MaxCost:            1 << 30, // Maximum cost of cache (1GB)
		Metrics:            false,   // Metrics are off by default
		NumCounters:        1e7,     // Number of keys to track frequency of (10M)
	}
}

// loadRistrettoClient will load the cache client (ristretto)
func loadRistrettoClient(
	ctx context.Context,
	config *ristretto.Config,
	newRelicEnabled bool,
) (*ristretto.Cache, error) {

	// Return silently
	if config == nil {
		return nil, ErrInvalidRistrettoConfig
	}

	// If NewRelic is enabled
	if newRelicEnabled {
		if txn := newrelic.FromContext(ctx); txn != nil {
			defer txn.StartSegment("load_ristretto_client").End()
		}
	}

	// Attempt to create the client
	client, err := ristretto.NewCache(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// writeLockRistretto will write a lock record into memory using a secret and expiration
//
// ttl is in seconds
func writeLockRistretto(ristrettoClient *ristretto.Cache, lockKey, secret string, cost, ttl int64) (bool, error) {

	// Test the key and secret
	if err := validateLockValues(lockKey, secret); err != nil {
		return false, err
	}

	// Try to get an existing lock (if it fails, make a new lock)
	data, ok := ristrettoClient.Get(lockKey)
	if !ok { // No lock found
		return ristrettoSet(ristrettoClient, lockKey, secret, cost, ttl)
	}

	// Check secret
	if data.(string) != secret { // Secret mismatch (lock exists with different secret)
		return false, cache.ErrLockMismatch
	}

	// Same secret / lock again?
	return ristrettoSet(ristrettoClient, lockKey, secret, cost, ttl)
}

// releaseLockRistretto will attempt to release a lock if it exists and matches the given secret
func releaseLockRistretto(ristrettoClient *ristretto.Cache, lockKey, secret string) (bool, error) {

	// Test the key and secret
	if err := validateLockValues(lockKey, secret); err != nil {
		return false, err
	}

	// Try to get an existing lock (if it fails, lock does not exist)
	data, ok := ristrettoClient.Get(lockKey)
	if !ok { // No lock found
		return true, nil
	}

	// Check secret if found
	if data.(string) == secret { // If it matches, remove the key
		ristrettoClient.Del(lockKey)
		ristrettoClient.Wait()
		return true, nil
	}

	// Key found does not match the secret, do not remove
	return false, cache.ErrLockMismatch
}

// ristrettoSet will set a key in ristretto and check for the error
//
// ttl is in seconds
func ristrettoSet(ristrettoClient *ristretto.Cache, key string, value interface{}, cost, ttl int64) (bool, error) {
	if success := ristrettoClient.SetWithTTL(
		key, value, cost, time.Second*time.Duration(ttl),
	); !success {
		return false, ErrRistrettoSetFailed
	}
	ristrettoClient.Wait()
	return true, nil
}
