package chainstate

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mrz1836/go-mattercloud"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonicpow/go-minercraft"
)

// todo: finish these unit tests!

// TestClient_Close will test the method Close()
func TestClient_Close(t *testing.T) {
	// finish test
}

// TestClient_Debug will test the method Debug()
func TestClient_Debug(t *testing.T) {
	// finish test
}

// TestClient_IsDebug will test the method IsDebug()
func TestClient_IsDebug(t *testing.T) {
	// finish test
}

// TestClient_DebugLog will test the method DebugLog()
func TestClient_DebugLog(t *testing.T) {
	// finish test
}

// TestNewClient will test the method NewClient()
func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("basic defaults", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, false, c.IsDebug())
		assert.Equal(t, MainNet, c.Network())
		assert.Nil(t, c.HTTPClient())
		assert.NotNil(t, c.MatterCloud())
		assert.NotNil(t, c.WhatsOnChain())
		assert.NotNil(t, c.Minercraft())
	})

	t.Run("custom http client", func(t *testing.T) {
		customClient := &http.Client{}
		c, err := NewClient(
			context.Background(),
			WithHTTPClient(customClient),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.HTTPClient())
		assert.Equal(t, customClient, c.HTTPClient())
	})

	t.Run("custom whatsonchain client", func(t *testing.T) {
		customClient := whatsonchain.NewClient(
			MainNet.WhatsOnChain(), whatsonchain.ClientDefaultOptions(), nil,
		)
		require.NotNil(t, customClient)
		c, err := NewClient(
			context.Background(),
			WithWhatsOnChain(customClient),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.WhatsOnChain())
		assert.Equal(t, customClient, c.WhatsOnChain())
	})

	t.Run("custom mattercloud client", func(t *testing.T) {
		customClient, err := mattercloud.NewClient(
			testDummyKey, MainNet.MatterCloud(), mattercloud.ClientDefaultOptions(), nil,
		)
		require.NoError(t, err)
		require.NotNil(t, customClient)

		var c ClientInterface
		c, err = NewClient(
			context.Background(),
			WithMatterCloud(customClient),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.MatterCloud())
		assert.Equal(t, customClient, c.MatterCloud())
	})

	t.Run("custom matter cloud api key", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			WithMatterCloudAPIKey(testDummyKey),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.MatterCloud())
	})

	t.Run("custom whats on chain api key", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			WithWhatsOnChainAPIKey(testDummyKey),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.WhatsOnChain())
	})

	t.Run("custom minercraft client", func(t *testing.T) {
		customClient, err := minercraft.NewClient(
			minercraft.DefaultClientOptions(), nil, nil,
		)
		require.NoError(t, err)
		require.NotNil(t, customClient)

		var c ClientInterface
		c, err = NewClient(
			context.Background(),
			WithMinercraft(customClient),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.Minercraft())
		assert.Equal(t, customClient, c.Minercraft())
	})

	t.Run("custom list of broadcast miners", func(t *testing.T) {
		miners, _ := defaultMiners()
		c, err := NewClient(
			context.Background(),
			WithBroadcastMiners(miners),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, miners, c.BroadcastMiners())
	})

	t.Run("custom list of query miners", func(t *testing.T) {
		miners, _ := defaultMiners()
		c, err := NewClient(
			context.Background(),
			WithQueryMiners(miners),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, miners, c.QueryMiners())
	})

	t.Run("custom query timeout", func(t *testing.T) {
		timeout := 55 * time.Second
		c, err := NewClient(
			context.Background(),
			WithQueryTimeout(timeout),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, timeout, c.QueryTimeout())
	})

	t.Run("custom network - test", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			WithNetwork(TestNet),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, TestNet, c.Network())
	})

	t.Run("custom network - stn", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			WithNetwork(StressTestNet),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, StressTestNet, c.Network())
	})
}
