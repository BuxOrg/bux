package cachestore

import (
	"context"

	"github.com/coocood/freecache"
	"github.com/mrz1836/go-cache"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type (

	// Client is the client (configuration)
	Client struct {
		options *clientOptions
	}

	// clientOptions holds all the configuration for the client
	clientOptions struct {
		debug           bool             // For extra logs and additional debug information
		engine          Engine           // Cachestore engine (redis or mcache)
		freecache       *freecache.Cache // Driver (client) for local in-memory storage
		logger          Logger           // Internal logging
		newRelicEnabled bool             // If NewRelic is enabled (parent application)
		redis           *cache.Client    // Current redis client (read & write)
		redisConfig     *RedisConfig     // Configuration for a new redis client
	}
)

// NewClient creates a new client for all CacheStore functionality
//
// If no options are given, it will use the defaultClientOptions()
// ctx may contain a NewRelic txn (or one will be created)
func NewClient(ctx context.Context, opts ...ClientOps) (ClientInterface, error) {

	// Create a new client with defaults
	client := &Client{options: defaultClientOptions()}

	// Overwrite defaults with any set by user
	for _, opt := range opts {
		opt(client.options)
	}

	// Set logger if not set
	if client.options.logger == nil {
		client.options.logger = newLogger()
	}

	// EMPTY! Engine was NOT set, show warning and use in-memory cache
	if client.Engine().IsEmpty() {
		client.options.logger.Warn(ctx, "cachestore engine was not set, using in-memory FreeCache")
		client.options.engine = FreeCache
	}

	// Use NewRelic if it's enabled (use existing txn if found on ctx)
	ctx = client.options.getTxnCtx(ctx)

	// Load cache based on engine
	if client.Engine() == Redis {

		// Only if we don't already have an existing client
		if client.options.redis == nil {
			var err error
			if client.options.redis, err = loadRedisClient(
				ctx, client.options.redisConfig, client.options.newRelicEnabled,
			); err != nil {
				return nil, err
			}
		}
	} else if client.Engine() == FreeCache {

		// Only if we don't already have an existing client
		if client.options.freecache == nil {
			client.options.freecache = loadFreeCache(DefaultCacheSize, DefaultGCPercent)
		}
	}

	// Return the client
	return client, nil
}

// Close will close the client and any open connections
func (c *Client) Close(ctx context.Context) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("close_cachestore").End()
	}
	if c != nil && c.options != nil {
		if c.Engine() == Redis {
			if c.options.redis != nil {
				c.options.redis.Close()
			}
			c.options.redis = nil
		} else if c.Engine() == FreeCache {
			if c.options.freecache != nil {
				c.options.freecache.Clear()
			}
			c.options.freecache = nil
		}
		c.options.engine = Empty
	}
}

// Debug will set the debug flag
func (c *Client) Debug(on bool) {
	c.options.debug = on
}

// IsDebug will return if debugging is enabled
func (c *Client) IsDebug() bool {
	return c.options.debug
}

// IsNewRelicEnabled will return if new relic is enabled
func (c *Client) IsNewRelicEnabled() bool {
	return c.options.newRelicEnabled
}

// Engine will return the engine that is set
func (c *Client) Engine() Engine {
	return c.options.engine
}

// Redis will return the Redis client if found
func (c *Client) Redis() *cache.Client {
	return c.options.redis
}

// RedisConfig will return the Redis config client if found
func (c *Client) RedisConfig() *RedisConfig {
	return c.options.redisConfig
}

// FreeCache will return the FreeCache client if found
func (c *Client) FreeCache() *freecache.Cache {
	return c.options.freecache
}

// EmptyCache will empty the cache entirely
//
// CAUTION: this will dump all the stored cache
func (c *Client) EmptyCache(ctx context.Context) error {
	if c.Engine() == Redis && c.options.redis != nil {
		return cache.DestroyCache(ctx, c.options.redis)
	} else if c.options.freecache != nil {
		c.options.freecache.Clear()
	}
	return nil
}
