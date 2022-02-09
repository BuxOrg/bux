package cachestore

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/mrz1836/go-cache"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// loadRedisClient will load the cache client (redis)
func loadRedisClient(
	ctx context.Context,
	config *RedisConfig,
	newRelicEnabled bool,
) (*cache.Client, error) {

	// Check for a config
	if config == nil || config.URL == "" {
		return nil, ErrInvalidRedisConfig
	}

	// If NewRelic is enabled
	if newRelicEnabled {
		if txn := newrelic.FromContext(ctx); txn != nil {
			segment := txn.StartSegment("load_redis_client")
			segment.AddAttribute("url", config.URL)
			defer segment.End()
		}
	}

	// Attempt to create the client
	client, err := cache.Connect(
		ctx,
		config.URL,
		config.MaxActiveConnections,
		config.MaxIdleConnections,
		config.MaxConnectionLifetime,
		config.MaxIdleTimeout,
		config.DependencyMode,
		newRelicEnabled,
		redis.DialUseTLS(config.UseTLS),
	)
	if err != nil {
		return nil, err
	}

	// Test the connection if DependencyMode mode is off (no connection tested)
	if !config.DependencyMode { // Fire a ping to make sure it works!
		if err = cache.Ping(ctx, client); err != nil {
			return nil, err
		}
	}
	return client, nil
}
