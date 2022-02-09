package cachestore

import (
	"context"

	"github.com/OrlovEvgeny/go-mcache"
	"github.com/dgraph-io/ristretto"
	"github.com/mrz1836/go-cache"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// ClientOps allow functional options to be supplied
// that overwrite default client options.
type ClientOps func(c *clientOptions)

// defaultClientOptions will return an clientOptions struct with the default settings
//
// Useful for starting with the default and then modifying as needed
func defaultClientOptions() *clientOptions {

	// Set the default options
	return &clientOptions{
		debug:           false,
		engine:          Empty,
		newRelicEnabled: false,
		redisConfig:     &RedisConfig{},
		ristrettoConfig: &ristretto.Config{},
	}
}

// getTxnCtx will check for an existing transaction
func (c *clientOptions) getTxnCtx(ctx context.Context) context.Context {
	if c.newRelicEnabled {
		txn := newrelic.FromContext(ctx)
		if txn != nil {
			ctx = newrelic.NewContext(ctx, txn)
		}
	}
	return ctx
}

// WithNewRelic will enable the NewRelic wrapper
func WithNewRelic() ClientOps {
	return func(c *clientOptions) {
		c.newRelicEnabled = true
	}
}

// WithDebugging will enable debugging mode
func WithDebugging() ClientOps {
	return func(c *clientOptions) {
		c.debug = true
	}
}

// WithRedis will set the redis configuration
func WithRedis(redisConfig *RedisConfig) ClientOps {
	return func(c *clientOptions) {

		// Don't panic if nil is passed
		if redisConfig == nil {
			return
		}

		// Set the config and engine
		c.redisConfig = redisConfig
		c.engine = Redis
		c.redis = nil // If you load via config, remove the connection

		// Set any defaults
		if c.redisConfig.MaxIdleTimeout.String() == emptyTimeDuration {
			c.redisConfig.MaxIdleTimeout = DefaultRedisMaxIdleTimeout
		}
	}
}

// WithRedisConnection will set an existing redis connection (read & write)
func WithRedisConnection(redisClient *cache.Client) ClientOps {
	return func(c *clientOptions) {
		if redisClient != nil {
			c.redis = redisClient
			c.engine = Redis
			c.redisConfig = nil // If you load an existing connection, config is not needed
		}
	}
}

// WithMcache will set the cache to local memory using mCache
func WithMcache() ClientOps {
	return func(c *clientOptions) {
		c.mCache = mcache.New()
		c.engine = MCache
	}
}

// WithRistretto will set the cache to local in-memory using Ristretto
func WithRistretto(config *ristretto.Config) ClientOps {
	return func(c *clientOptions) {

		// Don't panic if nil is passed
		if config == nil {
			return
		}

		// Set the config and engine
		c.ristrettoConfig = config
		c.engine = Ristretto
		c.ristretto = nil // If you load via config, remove the connection
	}
}

// WithRistrettoConnection will set an existing ristretto connection (read & write)
func WithRistrettoConnection(ristrettoClient *ristretto.Cache) ClientOps {
	return func(c *clientOptions) {
		if ristrettoClient != nil {
			c.ristretto = ristrettoClient
			c.engine = Ristretto
			c.ristrettoConfig = nil // If you load an existing connection, config is not needed
		}
	}
}

// WithMcacheConnection will set the cache to a current mCache driver connection
func WithMcacheConnection(driver *mcache.CacheDriver) ClientOps {
	return func(c *clientOptions) {
		if driver != nil {
			c.mCache = driver
			c.engine = MCache
		}
	}
}
