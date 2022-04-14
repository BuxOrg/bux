package cachestore

import (
	"context"
	"time"

	"github.com/coocood/freecache"
	"github.com/mrz1836/go-cache"
)

// LockService are the locking related methods
type LockService interface {
	ReleaseLock(ctx context.Context, lockKey, secret string) (bool, error)
	WaitWriteLock(ctx context.Context, lockKey string, ttl, ttw int64) (string, error)
	WriteLock(ctx context.Context, lockKey string, ttl int64) (string, error)
}

// CacheService are the cache related methods
type CacheService interface {
	Delete(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (string, error)
	GetModel(ctx context.Context, key string, model interface{}) error
	Set(ctx context.Context, key string, value interface{}, dependencies ...string) error
	SetModel(ctx context.Context, key string, model interface{}, ttl time.Duration, dependencies ...string) error
}

// ClientInterface is the cachestore interface
type ClientInterface interface {
	CacheService
	LockService
	Close(ctx context.Context)
	Debug(on bool)
	EmptyCache(ctx context.Context) error
	Engine() Engine
	FreeCache() *freecache.Cache
	IsDebug() bool
	IsNewRelicEnabled() bool
	Redis() *cache.Client
	RedisConfig() *RedisConfig
}
