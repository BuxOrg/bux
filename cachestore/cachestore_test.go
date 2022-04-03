package cachestore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/BuxOrg/bux/tester"
	"github.com/BuxOrg/bux/utils"
	"github.com/alicebob/miniredis/v2"
	"github.com/mrz1836/go-cache"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testIdleTimeout          = 240 * time.Second
	testKey                  = "test-key"
	testLocalConnectionURL   = RedisPrefix + "localhost:6379"
	testMaxActiveConnections = 0
	testMaxConnLifetime      = 60 * time.Second
	testMaxIdleConnections   = 10
	testValue                = "test-value"
	testAppName              = "test-app"
	testTxn                  = "test-txn"
)

// genericStruct is an example struct for testing
type genericStruct struct {
	StringField string
	IntField    int
	BoolField   bool
	FloatField  float64
}

// newMockRedisClient will create a new redis mock client
func newMockRedisClient(t *testing.T) (ClientInterface, *redigomock.Conn) {
	redisClient, conn := tester.LoadMockRedis(
		testIdleTimeout, testMaxConnLifetime, testMaxActiveConnections, testMaxIdleConnections,
	)
	require.NotNil(t, redisClient)
	require.NotNil(t, conn)

	c, err := NewClient(context.Background(), WithRedisConnection(redisClient))
	require.NotNil(t, c)
	require.NoError(t, err)
	return c, conn
}

// cacheTestCase is the test case struct
type cacheTestCase struct {
	engine Engine
	name   string
	opts   ClientOps
	redis  *miniredis.Miniredis
}

// FastForward changes the TTL for in-memory cache
func (c cacheTestCase) FastForward(duration time.Duration) {
	if c.engine == Redis && c.redis != nil {
		c.redis.FastForward(duration)
	}
}

// TestClient_SetRedis will test the method Set() and Get()
func TestClient_SetRedis(t *testing.T) {
	t.Run("[redis] [mocked] - valid get/set using redis", func(t *testing.T) {
		ctx := context.Background()
		client, conn := newMockRedisClient(t)

		require.NotNil(t, client)
		require.NotNil(t, conn)

		fees := "512"

		// Set command
		setCmd := conn.Command(cache.SetCommand, testKey, fees).Expect(fees)
		err := client.Set(ctx, testKey, fees)
		require.NoError(t, err)
		assert.Equal(t, true, setCmd.Called)

		// Get command
		getCmd := conn.Command(cache.GetCommand, testKey).Expect(fees)
		getFees, err2 := client.Get(ctx, testKey)
		require.NoError(t, err2)
		assert.Equal(t, true, getCmd.Called)

		assert.Equal(t, fees, getFees)
	})

	t.Run("[redis] [in-memory] valid get/set using redis", func(t *testing.T) {

		r := loadRedisInMemoryClient(t)
		require.NotNil(t, r)
		ctx := context.Background()

		c, err := NewClient(ctx,
			WithRedis(&RedisConfig{
				URL: r.Addr(),
			}),
			WithDebugging(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)

		defer func() {
			_ = c.EmptyCache(context.Background())
		}()

		fees := "512"

		// Set command
		err = c.Set(ctx, testKey, fees)
		require.NoError(t, err)

		// Get command
		getFees, err2 := c.Get(ctx, testKey)
		require.NoError(t, err2)

		assert.Equal(t, fees, getFees)
	})
}

// TestClient_SetModelRedis will test the method SetModel()
func TestClient_SetModelRedis(t *testing.T) {
	t.Run("[redis] [in-memory] valid set/get model struct", func(t *testing.T) {
		r := loadRedisInMemoryClient(t)
		require.NotNil(t, r)

		ctx := context.Background()

		c, err := NewClient(ctx,
			WithRedis(&RedisConfig{
				URL: r.Addr(),
			}),
			WithDebugging(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)

		defer func() {
			_ = c.EmptyCache(context.Background())
		}()

		fee := &utils.FeeUnit{
			Satoshis: 500,
			Bytes:    1000,
		}

		err = c.SetModel(ctx, testKey, fee, 1*time.Minute)
		require.NoError(t, err)

		getFee := new(utils.FeeUnit)
		err = c.GetModel(ctx, testKey, getFee)
		require.NoError(t, err)
		assert.Equal(t, 500, getFee.Satoshis)
		assert.Equal(t, 1000, getFee.Bytes)
	})
}

// TestClient_Get will test the method Get()
func TestClient_Get(t *testing.T) {

	testCases := getInMemoryTestCases(t)
	for _, testCase := range testCases {
		t.Run(testCase.name+" - empty key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			_, err = c.Get(context.Background(), "")
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - just spaces", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			_, err = c.Get(context.Background(), "   ")
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - key not found (nil)", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			defer func() {
				_ = c.EmptyCache(context.Background())
			}()

			err = c.Set(context.Background(), testKey+"-not-found", testValue)
			require.NoError(t, err)

			var val interface{}
			val, err = c.Get(context.Background(), testKey)
			require.NoError(t, err)
			assert.Equal(t, "", val)
		})

		t.Run(testCase.name+" - valid key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			defer func() {
				_ = c.EmptyCache(context.Background())
			}()

			err = c.Set(context.Background(), testKey, testValue)
			require.NoError(t, err)

			var val interface{}
			val, err = c.Get(context.Background(), testKey)
			require.NoError(t, err)
			assert.Equal(t, testValue, val.(string))
		})
	}

	t.Run("[redis] [mock] - empty key", func(t *testing.T) {
		c, _ := newMockRedisClient(t)

		_, err := c.Get(context.Background(), "")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("[redis] [mock] - valid key", func(t *testing.T) {
		c, conn := newMockRedisClient(t)

		// Mock the redis string
		setCmd := conn.Command(cache.SetCommand, testKey, testValue).Expect(testValue)

		err := c.Set(context.Background(), testKey, testValue)
		require.NoError(t, err)

		assert.Equal(t, true, setCmd.Called)

		// The main command to test
		getCmd := conn.Command(cache.GetCommand, testKey).Expect(testValue)

		var val interface{}
		val, err = c.Get(context.Background(), testKey)
		require.NoError(t, err)
		assert.Equal(t, testValue, val.(string))

		assert.Equal(t, true, getCmd.Called)
	})
}

// TestClient_Set will test the method Set()
func TestClient_Set(t *testing.T) {

	testCases := getInMemoryTestCases(t)
	for _, testCase := range testCases {
		t.Run(testCase.name+" - empty key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.Set(context.Background(), "", "")
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - just spaces", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.Set(context.Background(), "   ", "")
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - valid key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			defer func() {
				_ = c.EmptyCache(context.Background())
			}()

			err = c.Set(context.Background(), testKey, "")
			require.NoError(t, err)
		})

		t.Run(testCase.name+" - valid key, with leading and trailing spaces", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			defer func() {
				_ = c.EmptyCache(context.Background())
			}()

			err = c.Set(context.Background(), " "+testKey+" ", testValue)
			require.NoError(t, err)

			var val interface{}
			val, err = c.Get(context.Background(), testKey)
			require.NoError(t, err)
			assert.Equal(t, testValue, val.(string))
		})
	}

	t.Run("[redis] [mock] - empty key", func(t *testing.T) {
		c, _ := newMockRedisClient(t)

		err := c.Set(context.Background(), "", "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("[redis] [mock] - valid key", func(t *testing.T) {
		c, conn := newMockRedisClient(t)

		// Mock the redis string
		setCmd := conn.Command(cache.SetCommand, testKey, testValue).Expect(testValue)

		err := c.Set(context.Background(), testKey, testValue)
		require.NoError(t, err)

		assert.Equal(t, true, setCmd.Called)
	})
}

// TestClient_Delete will test the method Delete()
func TestClient_Delete(t *testing.T) {

	testCases := getInMemoryTestCases(t)
	for _, testCase := range testCases {
		t.Run(testCase.name+" - empty key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.Delete(context.Background(), "")
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - just spaces", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.Delete(context.Background(), "   ")
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - valid key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			defer func() {
				_ = c.EmptyCache(context.Background())
			}()

			err = c.Set(context.Background(), testKey, "")
			require.NoError(t, err)

			err = c.Delete(context.Background(), testKey)
			require.NoError(t, err)

			var val string
			val, err = c.Get(context.Background(), testKey)
			require.NoError(t, err)
			assert.Equal(t, "", val)
		})

		t.Run(testCase.name+" - valid key, with leading and trailing spaces", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			defer func() {
				_ = c.EmptyCache(context.Background())
			}()

			err = c.Set(context.Background(), " "+testKey+" ", testValue)
			require.NoError(t, err)

			err = c.Delete(context.Background(), testKey)
			require.NoError(t, err)

			var val string
			val, err = c.Get(context.Background(), testKey)
			require.NoError(t, err)
			assert.Equal(t, "", val)
		})
	}

	t.Run("[redis] [mock] - empty key", func(t *testing.T) {
		c, _ := newMockRedisClient(t)

		err := c.Delete(context.Background(), "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("[redis] [mock] - valid key", func(t *testing.T) {
		c, conn := newMockRedisClient(t)

		// Mock the redis string
		delCmd := conn.Command(cache.DeleteCommand, testKey).Expect(1)

		err := c.Delete(context.Background(), testKey)
		require.NoError(t, err)

		assert.Equal(t, true, delCmd.Called)
	})
}

// TestClient_GetModel will test the method GetModel()
func TestClient_GetModel(t *testing.T) {

	testModel := &genericStruct{
		StringField: testValue,
		IntField:    123,
		BoolField:   true,
		FloatField:  12.34,
	}

	testCases := getInMemoryTestCases(t)
	for _, testCase := range testCases {
		t.Run(testCase.name+" - empty key", func(t *testing.T) {
			testModelEmpty := new(genericStruct)
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.GetModel(context.Background(), "", testModelEmpty)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - just spaces", func(t *testing.T) {
			testModelEmpty := new(genericStruct)
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.GetModel(context.Background(), "   ", testModelEmpty)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - valid key", func(t *testing.T) {
			testModelEmpty := new(genericStruct)
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			defer func() {
				_ = c.EmptyCache(context.Background())
			}()

			err = c.SetModel(context.Background(), testKey, testModel, 0)
			require.NoError(t, err)

			err = c.GetModel(context.Background(), testKey, testModelEmpty)
			require.NoError(t, err)
			assert.Equal(t, testModel.StringField, testModelEmpty.StringField)
			assert.Equal(t, testModel.IntField, testModelEmpty.IntField)
			assert.Equal(t, testModel.BoolField, testModelEmpty.BoolField)
			assert.Equal(t, testModel.FloatField, testModelEmpty.FloatField)
		})

		t.Run(testCase.name+" - record does not exist", func(t *testing.T) {
			testModelEmpty := new(genericStruct)
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.GetModel(context.Background(), testKey, testModelEmpty)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyNotFound)
			assert.NotEqual(t, testModel.StringField, testModelEmpty.StringField)
		})
	}

	t.Run("[redis] [mock] - empty key", func(t *testing.T) {
		testModelEmpty := new(genericStruct)
		c, _ := newMockRedisClient(t)

		err := c.GetModel(context.Background(), "", testModelEmpty)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyRequired)
	})

	t.Run("[redis] [mock] - record does not exist", func(t *testing.T) {
		testModelEmpty := new(genericStruct)
		c, conn := newMockRedisClient(t)

		getCmd := conn.Command(cache.GetCommand, testKey).Expect(nil)

		err := c.GetModel(context.Background(), testKey, testModelEmpty)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrKeyNotFound)
		assert.Equal(t, true, getCmd.Called)
	})

	t.Run("[redis] [mock] - record exists", func(t *testing.T) {
		testModelEmpty := new(genericStruct)
		c, conn := newMockRedisClient(t)

		responseBytes, err := json.Marshal(&testModel)
		require.NoError(t, err)

		setCmd := conn.Command(cache.SetCommand, testKey, string(responseBytes))

		err = c.SetModel(context.Background(), testKey, testModel, 0)
		require.NoError(t, err)
		assert.Equal(t, true, setCmd.Called)

		getCmd := conn.Command(cache.GetCommand, testKey).Expect(responseBytes)

		err = c.GetModel(context.Background(), testKey, testModelEmpty)
		require.NoError(t, err)
		assert.Equal(t, true, getCmd.Called)

		assert.Equal(t, testModel.StringField, testModelEmpty.StringField)
	})
}

// TestClient_SetModel will test the method SetModel()
func TestClient_SetModel(t *testing.T) {

	testModel := &genericStruct{
		StringField: testValue,
		IntField:    123,
		BoolField:   true,
		FloatField:  12.34,
	}

	testCases := getInMemoryTestCases(t)
	for _, testCase := range testCases {
		t.Run(testCase.name+" - empty key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.SetModel(context.Background(), "", testModel, 0)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - just spaces", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.Set(context.Background(), "   ", testModel)
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrKeyRequired)
		})

		t.Run(testCase.name+" - valid key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			defer func() {
				_ = c.EmptyCache(context.Background())
			}()

			err = c.SetModel(context.Background(), testKey, testModel, 0)
			require.NoError(t, err)
		})
	}

	t.Run("[redis] [mock] - valid key", func(t *testing.T) {
		c, conn := newMockRedisClient(t)

		responseBytes, err := json.Marshal(&testModel)
		require.NoError(t, err)

		setCmd := conn.Command(cache.SetCommand, testKey, string(responseBytes))

		err = c.SetModel(context.Background(), testKey, testModel, 0)
		require.NoError(t, err)

		assert.Equal(t, true, setCmd.Called)
	})
}

// getInMemoryTestCases will return all the cache engine test cases for in-memory testing
func getInMemoryTestCases(t *testing.T) (cases []cacheTestCase) {
	cases = []cacheTestCase{
		{
			name:   "[freecache] [in-memory]",
			engine: FreeCache,
			opts:   WithFreeCache(),
			redis:  nil,
		},
	}

	r := loadRedisInMemoryClient(t)
	require.NotNil(t, r)

	cases = append(cases, cacheTestCase{
		name:   "[redis] [in-memory]",
		engine: Redis,
		redis:  r,
		opts: WithRedis(&RedisConfig{
			MaxConnectionLifetime: DefaultRedisMaxIdleTimeout,
			MaxIdleTimeout:        DefaultRedisMaxIdleTimeout,
			URL:                   r.Addr(),
		}),
	})
	return
}
