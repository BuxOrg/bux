package cachestore

import (
	"errors"
)

// ErrKeyNotFound is returned when a record is not found for a given key
var ErrKeyNotFound = errors.New("key not found")

// ErrNoEngine is returned when there is no engine set (missing engine)
var ErrNoEngine = errors.New("cachestore engine is empty: choose redis or memory (IE: WithRedis())")

// ErrKeyRequired is returned when the key is empty (key->value)
var ErrKeyRequired = errors.New("key is empty and required")

// ErrSecretRequired is returned when the secret is empty (value)
var ErrSecretRequired = errors.New("secret is empty and required")

// ErrSecretGenerationFailed is the error if the secret failed to generate
var ErrSecretGenerationFailed = errors.New("failed generating secret")

// ErrLockCreateFailed is the error when creating a lock fails
var ErrLockCreateFailed = errors.New("failed creating cache lock")

// ErrLockExists is the error when trying to create a lock fails due to an existing lock
var ErrLockExists = errors.New("lock already exists with a different secret")

// ErrTTWCannotBeEmpty is when the TTW field is empty
var ErrTTWCannotBeEmpty = errors.New("the TTW value cannot be empty")

// ErrInvalidRedisConfig is when the redis config is missing or invalid
var ErrInvalidRedisConfig = errors.New("invalid redis config")
