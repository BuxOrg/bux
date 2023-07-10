package chainstate

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewTestClient returns a test client
func NewTestClient(ctx context.Context, t *testing.T, opts ...ClientOps) ClientInterface {
	// WithMinercraft() is at the beginning so that it is possible to override minercraft's value with opts
	opts = append([]ClientOps{WithMinercraft(&MinerCraftBase{})}, opts...)
	c, err := NewClient(
		ctx, append(opts, WithDebugging())...,
	)
	require.NoError(t, err)
	require.NotNil(t, c)
	return c
}

// TestQueryTransactionFastest tests the querying for a transaction and returns the fastest response
func TestQueryTransactionFastest(t *testing.T) {
	t.Run("no tx ID", func(t *testing.T) {
		ctx := context.Background()
		c, err := NewClient(ctx, WithMinercraft(&MinerCraftBase{}))
		require.NoError(t, err)

		_, err = c.QueryTransactionFastest(ctx, "", RequiredInMempool, 5*time.Second)
		require.Error(t, err)
	})

	t.Run("fastest query", func(t *testing.T) {
		ctx := context.Background()
		c, err := NewClient(ctx, WithMinercraft(&MinerCraftBase{}))
		require.NoError(t, err)

		var txInfo *TransactionInfo
		txInfo, err = c.QueryTransactionFastest(ctx, testTransactionID, RequiredInMempool, 5*time.Second)
		require.NoError(t, err)
		assert.NotNil(t, txInfo)
	})
}
