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

	"github.com/OrlovEvgeny/go-mcache"
	"github.com/gomodule/redigo/redis"
	"github.com/mrz1836/go-cache"
)

// Set will set a key->value using the current engine
//
// NOTE: redis only supports dependency keys at this time
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
	} else if c.Engine() == Ristretto {
		if !c.options.ristretto.Set(key, value, baseCostPerKey) {
			return ErrFailedToSet
		}
		c.options.ristretto.Wait()
		return nil
	}

	// mCache
	return c.options.mCache.Set(key, value, mcache.TTL_FOREVER)
}

// Get will return a value from a given key
//
// Redis will be an interface{} but really a string (empty string)
// mCache/ristretto will be an interface{} and usually a pointer (empty nil)
func (c *Client) Get(ctx context.Context, key string) (interface{}, error) {

	// Sanitize the key (trailing or leading spaces)
	key = strings.TrimSpace(key)

	// Require a key to be present
	if len(key) == 0 {
		return "", ErrKeyRequired
	}

	// Switch on the engine
	if c.Engine() == Redis {
		str, err := cache.Get(ctx, c.options.redis, key)
		if err != nil {
			return "", err
		}
		return str, nil
	} else if c.Engine() == Ristretto {
		value, found := c.options.ristretto.Get(key)
		if !found {
			return nil, nil
		}
		return value, nil
	} else if c.Engine() == MCache {
		if data, ok := c.options.mCache.Get(key); ok {
			return data, nil
		}
		return nil, nil
	}

	// Not found
	return nil, nil
}

// SetModel will set any model or struct (parsing Model->JSON (bytes))
//
// Model needs to be a pointer to a struct
// NOTE: redis only supports dependency keys at this time
func (c *Client) SetModel(ctx context.Context, key string, model interface{}, ttl time.Duration, dependencies ...string) error {

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

	// mCache (store the bytes)
	if c.Engine() == MCache {
		if ttl == 0 {
			ttl = mcache.TTL_FOREVER
		}
		return c.options.mCache.Set(key, responseBytes, ttl)
	}

	// Ristretto (store the bytes)
	if !c.options.ristretto.SetWithTTL(key, responseBytes, baseCostPerKey, ttl) {
		return ErrFailedToSet
	}
	c.options.ristretto.Wait()

	return nil
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
	} else if c.Engine() == Ristretto {
		if value, found := c.options.ristretto.Get(key); found {
			by := value.([]byte)

			// Sanity check to make sure there is a value to unmarshal
			if len(by) == 0 {
				return ErrKeyNotFound
			}

			return json.Unmarshal(by, &model)
		}
	} else if c.Engine() == MCache {
		if b, ok := c.options.mCache.Get(key); ok {
			by := b.([]byte)

			// Sanity check to make sure there is a value to unmarshal
			if len(by) == 0 {
				return ErrKeyNotFound
			}

			return json.Unmarshal(by, &model)
		}
	}

	// Not found
	return ErrKeyNotFound
}
