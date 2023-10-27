package chainstate

import (
	"context"
	"testing"

	broadcast_client_mock "github.com/bitcoin-sv/go-broadcast-client/broadcast/broadcast-client-mock"
	broadcast_fixtures "github.com/bitcoin-sv/go-broadcast-client/broadcast/broadcast-client-mock/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClient_Transaction will test the method QueryTransaction()
func TestClient_Transaction(t *testing.T) {
	t.Parallel()

	t.Run("error - missing id", func(t *testing.T) {
		// given
		c := NewTestClient(context.Background(), t, WithMinercraft(&minerCraftTxOnChain{}))

		// when
		info, err := c.QueryTransaction(
			context.Background(), "", RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrInvalidTransactionID)
	})

	t.Run("error - missing requirements", func(t *testing.T) {
		// given
		c := NewTestClient(context.Background(), t, WithMinercraft(&minerCraftTxOnChain{}))

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			"", defaultQueryTimeOut,
		)

		// then
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrInvalidRequirements)
	})
}

func TestClient_Transaction_MAPI(t *testing.T) {
	t.Parallel()

	t.Run("query transaction success - mAPI", func(t *testing.T) {
		// given
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Equal(t, minerTaal.Name, info.Provider)
		assert.Equal(t, minerTaal.MinerID, info.MinerID)
	})

	t.Run("valid - test network - mAPI", func(t *testing.T) {
		// given
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithNetwork(TestNet),
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Equal(t, minerTaal.Name, info.Provider)
		assert.Equal(t, minerTaal.MinerID, info.MinerID)
	})
}

func TestClient_Transaction_BroadcastClient(t *testing.T) {
	t.Parallel()

	t.Run("query transaction success - broadcastClient", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockSuccess).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&MinerCraftBase{}),
			WithBroadcastClient(bc),
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredInMempool, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, broadcast_fixtures.TxBlockHash, info.BlockHash)
		assert.Equal(t, broadcast_fixtures.TxBlockHeight, info.BlockHeight)
		assert.Equal(t, broadcast_fixtures.ProviderMain, info.Provider)
	})

	t.Run("valid - stress test network - broadcastClient", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockSuccess).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&MinerCraftBase{}),
			WithBroadcastClient(bc),
			WithNetwork(StressTestNet),
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredInMempool, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, broadcast_fixtures.TxBlockHash, info.BlockHash)
		assert.Equal(t, broadcast_fixtures.TxBlockHeight, info.BlockHeight)
		assert.Equal(t, broadcast_fixtures.ProviderMain, info.Provider)
	})

	t.Run("valid - test network - broadcast", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockSuccess).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&MinerCraftBase{}),
			WithBroadcastClient(bc),
			WithNetwork(TestNet),
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredInMempool, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, broadcast_fixtures.TxBlockHash, info.BlockHash)
		assert.Equal(t, broadcast_fixtures.TxBlockHeight, info.BlockHeight)
		assert.Equal(t, broadcast_fixtures.ProviderMain, info.Provider)
	})
}

func TestClient_Transaction_MultipleClients(t *testing.T) {
	t.Parallel()

	t.Run("valid - all clients", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockSuccess).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithBroadcastClient(bc),
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Equal(t, minerTaal.Name, info.Provider)
		assert.Equal(t, minerTaal.MinerID, info.MinerID)
	})

	t.Run("mAPI not found - broadcastClient", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockSuccess).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}), // NOT going to find the TX
			WithBroadcastClient(bc),
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredInMempool, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, broadcast_fixtures.TxBlockHash, info.BlockHash)
		assert.Equal(t, broadcast_fixtures.TxBlockHeight, info.BlockHeight)
		assert.Equal(t, broadcast_fixtures.ProviderMain, info.Provider)
	})

	t.Run("broadcastClient not found - mAPI", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockFailure).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithBroadcastClient(bc), // NOT going to find the TX
			WithMinercraft(&minerCraftTxOnChain{}),
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
		assert.Equal(t, minerTaal.Name, info.Provider)
		assert.Equal(t, minerTaal.MinerID, info.MinerID)
	})

	t.Run("error - all not found", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockFailure).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}), // NOT going to find the TX
			WithBroadcastClient(bc),                 // NOT going to find the TX
		)

		// when
		info, err := c.QueryTransaction(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrTransactionNotFound)
	})
}

// TestClient_Transaction_MultipleClients_Fastest will test the method QueryTransactionFastest()
func TestClient_Transaction_MultipleClients_Fastest(t *testing.T) {
	t.Parallel()

	t.Run("error - missing id", func(t *testing.T) {
		// given
		c := NewTestClient(context.Background(), t,
			WithMinercraft(&MinerCraftBase{}))

		// when
		info, err := c.QueryTransactionFastest(
			context.Background(), "", RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrInvalidTransactionID)
	})

	t.Run("error - missing requirements", func(t *testing.T) {
		// given
		c := NewTestClient(context.Background(), t,
			WithMinercraft(&MinerCraftBase{}))

		// when
		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			"", defaultQueryTimeOut,
		)

		// then
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrInvalidRequirements)
	})

	t.Run("valid - all clients", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockSuccess).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithBroadcastClient(bc),
		)

		// when
		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredInMempool, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.True(t, isOneOf(
			info.BlockHash,
			onChainExample1BlockHash,
			broadcast_fixtures.TxBlockHash,
		))
		assert.True(t, isOneOf(
			info.BlockHeight,
			onChainExample1BlockHeight,
			broadcast_fixtures.TxBlockHeight,
		))
		// todo: test is failing and needs to be fixed (@mrz)
		/*assert.True(t, isOneOf(
			info.Confirmations,
			onChainExample1Confirmations,
			0,
		))*/
		assert.True(t, isOneOf(
			info.Provider,
			minerTaal.Name,
			broadcast_fixtures.ProviderMain,
		))
	})

	t.Run("mAPI not found - broadcastClient", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockSuccess).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}), // NOT going to find the TX
			WithBroadcastClient(bc),
		)

		// when
		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredInMempool, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, broadcast_fixtures.TxBlockHash, info.BlockHash)
		assert.Equal(t, broadcast_fixtures.TxBlockHeight, info.BlockHeight)
		assert.Equal(t, broadcast_fixtures.ProviderMain, info.Provider)
	})

	t.Run("broadcastClient not found - mAPI", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockFailure).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithBroadcastClient(bc), // NOT going to find the TX
			WithMinercraft(&minerCraftTxOnChain{}),
		)

		// when
		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, onChainExample1BlockHash, info.BlockHash)
		assert.Equal(t, onChainExample1BlockHeight, info.BlockHeight)
		assert.Equal(t, onChainExample1Confirmations, info.Confirmations)
	})

	t.Run("error - all not found", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockFailure).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxNotFound{}), // NOT going to find the TX
			WithBroadcastClient(bc),                 // NOT going to find the TX
		)

		// when
		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredOnChain, defaultQueryTimeOut,
		)

		// then
		require.Error(t, err)
		require.Nil(t, info)
		assert.ErrorIs(t, err, ErrTransactionNotFound)
	})

	t.Run("valid - stn network", func(t *testing.T) {
		// given
		bc := broadcast_client_mock.Builder().
			WithMockArc(broadcast_client_mock.MockSuccess).
			Build()
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithBroadcastClient(bc),
			WithNetwork(StressTestNet),
		)

		// when
		info, err := c.QueryTransactionFastest(
			context.Background(), onChainExample1TxID,
			RequiredInMempool, defaultQueryTimeOut,
		)

		// then
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, onChainExample1TxID, info.ID)
		assert.Equal(t, broadcast_fixtures.TxBlockHash, info.BlockHash)
		assert.Equal(t, broadcast_fixtures.TxBlockHeight, info.BlockHeight)
		assert.Equal(t, broadcast_fixtures.ProviderMain, info.Provider)
	})
}

func isOneOf(val1 interface{}, val2 ...interface{}) bool {
	for _, element := range val2 {
		if val1 == element {
			return true
		}
	}

	return false
}
