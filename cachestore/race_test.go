//go:build !race
// +build !race

package cachestore

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/BuxOrg/bux/tester"
	"github.com/mrz1836/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_Get will test the method Get()
func TestClient_Get(t *testing.T) {

	for _, testCase := range cacheTestCases {
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

		t.Run(testCase.name+" - valid key", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.Set(context.Background(), testKey, testValue)
			require.NoError(t, err)

			var val interface{}
			val, err = c.Get(context.Background(), testKey)
			require.NoError(t, err)
			assert.Equal(t, testValue, val.(string))
		})

		t.Run(testCase.name+" - key not found (nil)", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

			err = c.Set(context.Background(), testKey+"-not-found", testValue)
			require.NoError(t, err)

			var val interface{}
			val, err = c.Get(context.Background(), testKey)
			require.NoError(t, err)
			assert.Nil(t, val)
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

	for _, testCase := range cacheTestCases {
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

			err = c.Set(context.Background(), testKey, "")
			require.NoError(t, err)
		})

		t.Run(testCase.name+" - valid key, with leading and trailing spaces", func(t *testing.T) {
			c, err := NewClient(context.Background(), testCase.opts)
			require.NotNil(t, c)
			require.NoError(t, err)

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

	t.Run("[redis] [real] - valid key", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test: redis is required")
		}

		redisClient, _, err := tester.LoadRealRedis(
			testLocalConnectionURL, testIdleTimeout, testMaxConnLifetime,
			testMaxActiveConnections, testMaxIdleConnections, true, false,
		)
		require.NotNil(t, redisClient)
		require.NoError(t, err)

		var c ClientInterface
		c, err = NewClient(context.Background(),
			WithRedisConnection(redisClient),
		)
		require.NotNil(t, c)
		require.NoError(t, err)

		err = c.Set(context.Background(), testKey, testValue)
		require.NoError(t, err)

		var val interface{}
		val, err = c.Get(context.Background(), testKey)
		require.NoError(t, err)
		assert.Equal(t, testValue, val.(string))
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

	for _, testCase := range cacheTestCases {
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

	for _, testCase := range cacheTestCases {
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
