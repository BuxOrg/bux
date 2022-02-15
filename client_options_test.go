package bux

import (
	"context"
	"testing"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/tester"
	"github.com/OrlovEvgeny/go-mcache"
	"github.com/dgraph-io/ristretto"
	"github.com/go-redis/redis/v8"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonicpow/go-paymail"
)

// todo: finish unit tests!

// TestNewRelicOptions will test the method enable()
func Test_newRelicOptions_enable(t *testing.T) {
	t.Parallel()

	t.Run("enable with valid app", func(t *testing.T) {
		app, err := tester.GetNewRelicApp(defaultNewRelicApp)
		require.NoError(t, err)
		require.NotNil(t, app)

		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithNewRelic(app))

		var tc ClientInterface
		tc, err = NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			opts...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		tc.EnableNewRelic()
		assert.Equal(t, true, tc.IsNewRelicEnabled())
	})

	t.Run("enable with invalid app", func(t *testing.T) {
		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithNewRelic(nil))

		tc, err := NewClient(tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx), opts...)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		tc.EnableNewRelic()
		assert.Equal(t, false, tc.IsNewRelicEnabled())
	})
}

// Test_newRelicOptions_getOrStartTxn will test the method getOrStartTxn()
func Test_newRelicOptions_getOrStartTxn(t *testing.T) {
	t.Parallel()

	t.Run("Get a valid ctx and txn", func(t *testing.T) {
		app, err := tester.GetNewRelicApp(defaultNewRelicApp)
		require.NoError(t, err)
		require.NotNil(t, app)

		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithNewRelic(app))

		var tc ClientInterface
		tc, err = NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			opts...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		ctx := tc.GetOrStartTxn(context.Background(), "test-name")
		assert.NotNil(t, ctx)

		txn := newrelic.FromContext(ctx)
		assert.NotNil(t, txn)
	})

	t.Run("invalid ctx and txn", func(t *testing.T) {
		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithNewRelic(nil))

		tc, err := NewClient(tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx), opts...)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		ctx := tc.GetOrStartTxn(context.Background(), "test-name")
		assert.NotNil(t, ctx)

		txn := newrelic.FromContext(ctx)
		assert.Nil(t, txn)
	})
}

// TestClient_defaultModelOptions will test the method DefaultModelOptions()
func TestClient_defaultModelOptions(t *testing.T) {
	t.Parallel()

	t.Run("default options", func(t *testing.T) {
		dco := defaultClientOptions()
		require.NotNil(t, dco)

		require.NotNil(t, dco.cacheStore)
		require.Nil(t, dco.cacheStore.ClientInterface)
		require.NotNil(t, dco.cacheStore.options)
		assert.Equal(t, 0, len(dco.cacheStore.options))

		require.NotNil(t, dco.dataStore)
		require.Nil(t, dco.dataStore.ClientInterface)
		require.NotNil(t, dco.dataStore.options)
		assert.Equal(t, 0, len(dco.dataStore.options))

		require.NotNil(t, dco.newRelic)

		require.NotNil(t, dco.paymail)

		assert.Equal(t, defaultUserAgent, dco.userAgent)

		require.NotNil(t, dco.taskManager)

		assert.Equal(t, true, dco.itc)

		assert.Nil(t, dco.logger)
	})
}

// TestWithUserAgent will test the method WithUserAgent()
func TestWithUserAgent(t *testing.T) {
	t.Parallel()

	t.Run("empty user agent", func(t *testing.T) {
		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithUserAgent(""))

		tc, err := NewClient(tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx), opts...)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		assert.NotEqual(t, "", tc.UserAgent())
		assert.Equal(t, defaultUserAgent, tc.UserAgent())
	})

	t.Run("custom user agent", func(t *testing.T) {
		customAgent := "custom-user-agent"

		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithUserAgent(customAgent))

		tc, err := NewClient(tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx), opts...)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		assert.NotEqual(t, defaultUserAgent, tc.UserAgent())
		assert.Equal(t, customAgent, tc.UserAgent())
	})
}

// TestWithNewRelic will test the method WithNewRelic()
func TestWithNewRelic(t *testing.T) {
	// finish test
}

// TestWithDebugging will test the method WithDebugging()
func TestWithDebugging(t *testing.T) {
	t.Parallel()

	t.Run("set debug (with cache and Datastore)", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			DefaultClientOpts(t, true, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		assert.Equal(t, true, tc.IsDebug())
		assert.Equal(t, true, tc.Cachestore().IsDebug())
		assert.Equal(t, true, tc.Datastore().IsDebug())
		assert.Equal(t, true, tc.Taskmanager().IsDebug())
	})
}

// TestWithMcache will test the method WithMcache()
func TestWithMcache(t *testing.T) {
	t.Parallel()

	t.Run("using memory cache", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithMcache(),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}))
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		cs := tc.Cachestore()
		require.NotNil(t, cs)
		assert.Equal(t, cachestore.MCache, cs.Engine())
	})
}

// TestWithMcacheConnection will test the method WithMcacheConnection()
func TestWithMcacheConnection(t *testing.T) {
	t.Parallel()

	t.Run("using a nil driver", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithMcacheConnection(nil),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}))
		require.Error(t, err)
		require.Nil(t, tc)
	})

	t.Run("using an existing connection", func(t *testing.T) {
		mem := mcache.New()

		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithMcacheConnection(mem),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}))
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		cs := tc.Cachestore()
		require.NotNil(t, cs)
		assert.Equal(t, cachestore.MCache, cs.Engine())
	})
}

// TestWithRistretto will test the method WithRistretto()
func TestWithRistretto(t *testing.T) {
	t.Parallel()

	t.Run("empty config", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithRistrettoConnection(nil),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}))
		require.Error(t, err)
		require.Nil(t, tc)
	})

	t.Run("using valid config", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			DefaultClientOpts(t, false, true)...)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		cs := tc.Cachestore()
		require.NotNil(t, cs)
		assert.Equal(t, cachestore.Ristretto, cs.Engine())
	})
}

// TestWithRistrettoConnection will test the method WithRistrettoConnection()
func TestWithRistrettoConnection(t *testing.T) {
	t.Parallel()

	t.Run("using a nil driver", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithRistrettoConnection(nil),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}))
		require.Error(t, err)
		require.Nil(t, tc)
	})

	t.Run("using an existing connection", func(t *testing.T) {
		mem, err := ristretto.NewCache(cachestore.DefaultRistrettoConfig())
		require.NoError(t, err)
		require.NotNil(t, mem)

		var tc ClientInterface
		tc, err = NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithTaskQ(taskmanager.DefaultTaskQConfig(testQueueName), taskmanager.FactoryMemory),
			WithRistrettoConnection(mem),
			WithSQLite(&datastore.SQLiteConfig{Shared: true}))
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		cs := tc.Cachestore()
		require.NotNil(t, cs)
		assert.Equal(t, cachestore.Ristretto, cs.Engine())
	})
}

// TestWithRedis will test the method WithRedis()
func TestWithRedis(t *testing.T) {
	t.Run("using valid config", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithTaskQ(taskmanager.DefaultTaskQConfig(tester.RandomTablePrefix(t)), taskmanager.FactoryMemory),
			WithRedis(&cachestore.RedisConfig{
				URL: cachestore.RedisPrefix + "localhost:6379",
			}),
			WithSQLite(tester.SQLiteTestConfig(t, false, true)),
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		cs := tc.Cachestore()
		require.NotNil(t, cs)
		assert.Equal(t, cachestore.Redis, cs.Engine())
	})

	t.Run("missing redis prefix", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithTaskQ(taskmanager.DefaultTaskQConfig(tester.RandomTablePrefix(t)), taskmanager.FactoryMemory),
			WithRedis(&cachestore.RedisConfig{
				URL: "localhost:6379",
			}),
			WithSQLite(tester.SQLiteTestConfig(t, false, true)),
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		cs := tc.Cachestore()
		require.NotNil(t, cs)
		assert.Equal(t, cachestore.Redis, cs.Engine())
	})
}

// TestWithRedisConnection will test the method WithRedisConnection()
func TestWithRedisConnection(t *testing.T) {
	// finish test
}

// TestWithAutoMigrate will test the method WithAutoMigrate()
func TestWithAutoMigrate(t *testing.T) {
	// finish test
}

// TestWithSQLite will test the method WithSQLite()
func TestWithSQLite(t *testing.T) {
	// finish test
}

// TestWithSQL will test the method WithSQL()
func TestWithSQL(t *testing.T) {
	// finish test
}

// TestWithSQLConnection will test the method WithSQLConnection()
func TestWithSQLConnection(t *testing.T) {
	// finish test
}

// TestWithMongoDB will test the method WithMongoDB()
func TestWithMongoDB(t *testing.T) {
	// finish test
}

// TestWithMongoConnection will test the method WithMongoConnection()
func TestWithMongoConnection(t *testing.T) {
	// finish test
}

// TestWithPaymailClient will test the method WithPaymailClient()
func TestWithPaymailClient(t *testing.T) {
	t.Parallel()

	t.Run("using a nil driver, automatically makes paymail client", func(t *testing.T) {
		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithPaymailClient(nil))

		tc, err := NewClient(tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx), opts...)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		assert.NotNil(t, tc.PaymailClient())
	})

	t.Run("custom paymail client", func(t *testing.T) {
		p, err := paymail.NewClient()
		require.NoError(t, err)
		require.NotNil(t, p)

		opts := DefaultClientOpts(t, false, true)
		opts = append(opts, WithPaymailClient(p))

		var tc ClientInterface
		tc, err = NewClient(tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx), opts...)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		assert.NotNil(t, tc.PaymailClient())
		assert.Equal(t, p, tc.PaymailClient())
	})
}

// TestWithTaskQ will test the method WithTaskQ()
func TestWithTaskQ(t *testing.T) {
	t.Parallel()

	// todo: test cases where config is nil, or cannot load TaskQ

	t.Run("using taskq using memory", func(t *testing.T) {
		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			DefaultClientOpts(t, false, true)...,
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		tm := tc.Taskmanager()
		require.NotNil(t, tm)
		assert.Equal(t, taskmanager.TaskQ, tm.Engine())
		assert.Equal(t, taskmanager.FactoryMemory, tm.Factory())
	})

	t.Run("using taskq using redis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping live local redis tests")
		}

		tc, err := NewClient(
			tester.GetNewRelicCtx(t, defaultNewRelicApp, defaultNewRelicTx),
			WithTaskQUsingRedis(
				taskmanager.DefaultTaskQConfig(tester.RandomTablePrefix(t)),
				&redis.Options{
					Addr: "localhost:6379",
				},
			),
			WithRedis(&cachestore.RedisConfig{
				URL: cachestore.RedisPrefix + "localhost:6379",
			}),
			WithSQLite(tester.SQLiteTestConfig(t, false, true)),
		)
		require.NoError(t, err)
		require.NotNil(t, tc)
		defer CloseClient(context.Background(), t, tc)

		tm := tc.Taskmanager()
		require.NotNil(t, tm)
		assert.Equal(t, taskmanager.TaskQ, tm.Engine())
		assert.Equal(t, taskmanager.FactoryRedis, tm.Factory())
	})
}

// TestWithLogger will test the method WithLogger()
func TestWithLogger(t *testing.T) {
	// finish this!
}

// TestWithCustomTaskManager will test the method WithCustomTaskManager()
func TestWithCustomTaskManager(t *testing.T) {
	// finish this!
}

// TestWithCustomDatastore will test the method WithCustomDatastore()
func TestWithCustomDatastore(t *testing.T) {
	// finish this!
}

// TestWithCustomCachestore will test the method WithCustomCachestore()
func TestWithCustomCachestore(t *testing.T) {
	// finish this!
}

// TestWithCustomChainstate will test the method WithCustomChainstate()
func TestWithCustomChainstate(t *testing.T) {
	// finish this!
}

// TestWithModels will test the method WithModels()
func TestWithModels(t *testing.T) {
	// finish this!
}
