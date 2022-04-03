/*
Package cachestore is the caching (key->value) service abstraction layer
*/
package cachestore

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/coocood/freecache"
	"github.com/gomodule/redigo/redis"
	"github.com/mrz1836/go-cache"
)

// Set will set a key->value using the current engine
//
// NOTE: redis only supports dependency keys at this time
// Value should be used as a string for best results
func (c *Client) Set(ctx context.Context, key string, value interface{}, dependencies ...string) error {

	// Sanitize the key (trailing or leading spaces)
	key = strings.TrimSpace(key)

	// Require a key to be present
	if len(key) == 0 {
		return ErrKeyRequired
	}

	// Redis
	if c.Engine() == Redis {
		return cache.Set(ctx, c.options.redis, key, value, dependencies...)
	}

	// FreeCache
	return c.options.freecache.Set([]byte(key), []byte(value.(string)), 0)
}

// Get will return a value from a given key
//
// Redis will be an interface{} but really a string (empty string)
// mCache/ristretto will be an interface{} and usually a pointer (empty nil)
func (c *Client) Get(ctx context.Context, key string) (string, error) {

	// Sanitize the key (trailing or leading spaces)
	key = strings.TrimSpace(key)

	// Require a key to be present
	if len(key) == 0 {
		return "", ErrKeyRequired
	}

	// Switch on the engine
	if c.Engine() == Redis {
		str, err := cache.Get(ctx, c.options.redis, key)
		if err != nil && errors.Is(err, redis.ErrNil) {
			return "", nil
		} else if err != nil {
			return "", err
		}
		return str, nil
	}

	// Check using FreeCache
	data, err := c.options.freecache.Get([]byte(key))
	if err != nil && errors.Is(err, freecache.ErrNotFound) { // Ignore this error
		return "", nil
	} else if err != nil { // Real error getting the cache value
		return "", err
	}
	return string(data), nil
}

// Delete will remove a key from the cache
func (c *Client) Delete(ctx context.Context, key string) error {

	// Sanitize the key (trailing or leading spaces)
	key = strings.TrimSpace(key)

	// Require a key to be present
	if len(key) == 0 {
		return ErrKeyRequired
	}

	// Switch on the engine
	if c.Engine() == Redis {
		_, err := cache.DeleteWithoutDependency(ctx, c.options.redis, key)
		return err
	}

	// Use FreeCache
	_ = c.options.freecache.Del([]byte(key))
	return nil
}

// SetModel will set any model or struct (parsing Model->JSON (bytes))
//
// Model needs to be a pointer to a struct
// NOTE: redis only supports dependency keys at this time
func (c *Client) SetModel(ctx context.Context, key string, model interface{},
	ttl time.Duration, dependencies ...string) error {

	// Sanitize the key (trailing or leading spaces)
	key = strings.TrimSpace(key)

	// Require a key to be present
	if len(key) == 0 {
		return ErrKeyRequired
	}

	// Redis
	if c.Engine() == Redis {
		return cache.SetToJSON(ctx, c.options.redis, key, model, ttl, dependencies...)
	}

	// Parse into JSON
	responseBytes, err := json.Marshal(&model)
	if err != nil {
		return err
	}

	// FreeCache (store the bytes)
	return c.options.freecache.Set([]byte(key), responseBytes, int(ttl.Seconds()))
}

// GetModel will get a model (parsing JSON (bytes) -> Model)
//
// Model needs to be a pointer to a struct
func (c *Client) GetModel(ctx context.Context, key string, model interface{}) error {

	// Sanitize the key (trailing or leading spaces)
	key = strings.TrimSpace(key)

	// Require a key to be present
	if len(key) == 0 {
		return ErrKeyRequired
	}

	// Redis
	if c.Engine() == Redis {

		// Get the record as bytes
		b, err := cache.GetBytes(ctx, c.options.redis, key)
		if err != nil {
			if errors.Is(err, redis.ErrNil) {
				return ErrKeyNotFound
			}
			return err
		}

		// Sanity check to make sure there is a value to unmarshal
		if len(b) == 0 {
			return ErrKeyNotFound
		}

		return json.Unmarshal(b, &model)
	} else if c.Engine() == FreeCache {
		if b, err := c.options.freecache.Get([]byte(key)); err == nil && len(b) > 0 {
			return json.Unmarshal(b, &model)
		}
	}

	// Not found
	return ErrKeyNotFound
}
