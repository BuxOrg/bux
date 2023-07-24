package chainstate

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mrz1836/go-whatsonchain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonicpow/go-minercraft/v2"
)

// TestNewClient will test the method NewClient()
func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("basic defaults", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			WithMinercraft(&MinerCraftBase{}),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, false, c.IsDebug())
		assert.Equal(t, MainNet, c.Network())
		assert.Nil(t, c.HTTPClient())
		assert.NotNil(t, c.WhatsOnChain())
		assert.NotNil(t, c.Minercraft())
	})

	t.Run("custom http client", func(t *testing.T) {
		customClient := &http.Client{}
		c, err := NewClient(
			context.Background(),
			WithHTTPClient(customClient),
			WithMinercraft(&MinerCraftBase{}),
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
			WithMinercraft(&MinerCraftBase{}),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.WhatsOnChain())
		assert.Equal(t, customClient, c.WhatsOnChain())
	})

	t.Run("custom whats on chain api key", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			WithWhatsOnChainAPIKey(testDummyKey),
			WithMinercraft(&MinerCraftBase{}),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.NotNil(t, c.WhatsOnChain())
	})

	t.Run("custom minercraft client", func(t *testing.T) {
		customClient, err := minercraft.NewClient(
			minercraft.DefaultClientOptions(), nil, "", nil, nil,
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
			WithMinercraft(&MinerCraftBase{}),
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
			WithMinercraft(&MinerCraftBase{}),
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
			WithMinercraft(&MinerCraftBase{}),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, timeout, c.QueryTimeout())
	})

	t.Run("custom network - test", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			WithNetwork(TestNet),
			WithMinercraft(&MinerCraftBase{}),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, TestNet, c.Network())
	})

	t.Run("custom network - stn", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			WithNetwork(StressTestNet),
			WithMinercraft(&MinerCraftBase{}),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, StressTestNet, c.Network())
	})

	t.Run("unreacheble miners", func(t *testing.T) {
		_, err := NewClient(
			context.Background(),
			WithMinercraft(&minerCraftUnreachble{}),
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingBroadcastMiners)
	})
}
