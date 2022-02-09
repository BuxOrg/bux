package cachestore

import (
	"time"
)

const (
	// DefaultRedisMaxIdleTimeout is the default max timeout on an idle connection
	DefaultRedisMaxIdleTimeout = 240 * time.Second

	// Empty time duration for comparison
	emptyTimeDuration = "0s"

	// lockRetrySleepTime is in milliseconds
	lockRetrySleepTime = 10 * time.Millisecond

	// baseCostPerKey is the cost for each record
	baseCostPerKey = 1 // todo: this can be a variable set per request (in the future)
)

// RedisConfig is the configuration for the cache client (redis)
type RedisConfig struct {
	URL                   string        `json:"url" mapstructure:"url"`                                         // redis://localhost:6379
	MaxActiveConnections  int           `json:"max_active_connections" mapstructure:"max_active_connections"`   // 0
	MaxConnectionLifetime time.Duration `json:"max_connection_lifetime" mapstructure:"max_connection_lifetime"` // 0
	MaxIdleConnections    int           `json:"max_idle_connections" mapstructure:"max_idle_connections"`       // 10
	MaxIdleTimeout        time.Duration `json:"max_idle_timeout" mapstructure:"max_idle_timeout"`               // 240 * time.Second
	DependencyMode        bool          `json:"dependency_mode" mapstructure:"dependency_mode"`                 // false for digital ocean (not supported)
	UseTLS                bool          `json:"use_tls" mapstructure:"use_tls"`                                 // true for digital ocean (required)
}
