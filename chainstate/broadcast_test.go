package chainstate

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonicpow/go-minercraft/v2"
)

func Test_doesErrorContain(t *testing.T) {
	t.Run("valid contains", func(t *testing.T) {
		success := doesErrorContain("this is the test message", []string{"another", "test message"})
		assert.Equal(t, true, success)
	})

	t.Run("valid contains - equal case", func(t *testing.T) {
		success := doesErrorContain("this is the TEST message", []string{"another", "test message"})
		assert.Equal(t, true, success)
	})

	t.Run("does not contain", func(t *testing.T) {
		success := doesErrorContain("this is the test message", []string{"another", "nope"})
		assert.Equal(t, false, success)
	})
}

// TestClient_Broadcast_Success will test the method Broadcast()
func TestClient_Broadcast_Success(t *testing.T) {
	t.Parallel()

	t.Run("broadcast - success (mAPI)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftBroadcastSuccess{}),
			WithNowNodes(&nowNodesTxNotFound{}),         // Not found
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
		)
		providers, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		miners := strings.Split(providers, ",")

		assert.GreaterOrEqual(t, len(miners), 1)

		assert.True(t, containsAtLeastOneElement(miners, minercraft.MinerTaal, minercraft.MinerMempool, minercraft.MinerGorillaPool, minercraft.MinerMatterpool))

		assert.NotContains(t, miners, ProviderWhatsOnChain)
		assert.NotContains(t, miners, ProviderNowNodes)
	})

	t.Run("broadcast - success (WhatsOnChain)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithWhatsOnChain(&whatsOnChainBroadcastSuccess{}),
			WithNowNodes(&nowNodesTxNotFound{}),     // Not Found
			WithMinercraft(&minerCraftTxNotFound{}), // Not found
		)
		provider, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		assert.Equal(t, ProviderWhatsOnChain, provider)
	})

	t.Run("broadcast - success (NowNodes)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodesBroadcastSuccess{}),
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not found
		)
		provider, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		assert.Equal(t, ProviderNowNodes, provider)
	})

	t.Run("broadcast - success (NowNodes timeouts)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodesBroadcastTimeout{}),         // Timeout
			WithWhatsOnChain(&whatsOnChainBroadcastSuccess{}), // Success
			WithMinercraft(&minerCraftBroadcastSuccess{}),     // Success
		)
		providers, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		miners := strings.Split(providers, ",")

		assert.GreaterOrEqual(t, len(miners), 1)

		assert.True(t, containsAtLeastOneElement(miners, minercraft.MinerTaal, minercraft.MinerMempool, minercraft.MinerGorillaPool, minercraft.MinerMatterpool, ProviderWhatsOnChain))

		assert.NotContains(t, miners, ProviderNowNodes)
	})

	t.Run("broadcast - success (WhatsOnChain timeouts)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithWhatsOnChain(&whatsOnChainBroadcastTimeout{}), // Timeout
			WithNowNodes(&nowNodesBroadcastSuccess{}),         // Success
			WithMinercraft(&minerCraftBroadcastSuccess{}),     // Success
		)
		providers, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		miners := strings.Split(providers, ",")

		assert.GreaterOrEqual(t, len(miners), 1)

		assert.True(t, containsAtLeastOneElement(miners, minercraft.MinerTaal, minercraft.MinerMempool, minercraft.MinerGorillaPool, minercraft.MinerMatterpool, ProviderNowNodes))

		assert.NotContains(t, miners, ProviderWhatsOnChain)
	})

	t.Run("broadcast - success (mAPI timeouts)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftBroadcastTimeout{}),     // Timeout
			WithWhatsOnChain(&whatsOnChainBroadcastSuccess{}), // Success
			WithNowNodes(&nowNodesBroadcastSuccess{}),         // Success
		)
		providers, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		miners := strings.Split(providers, ",")

		assert.GreaterOrEqual(t, len(miners), 1)

		assert.True(t, containsAtLeastOneElement(miners, ProviderWhatsOnChain, ProviderNowNodes))

		assert.NotContains(t, miners, minercraft.MinerTaal)
		assert.NotContains(t, miners, minercraft.MinerMempool)
		assert.NotContains(t, miners, minercraft.MinerGorillaPool)
		assert.NotContains(t, miners, minercraft.MinerMatterpool)
	})
}

// TestClient_Broadcast_OnChain will test the method Broadcast()
func TestClient_Broadcast_OnChain(t *testing.T) {
	t.Parallel()

	t.Run("broadcast - tx already on-chain (mAPI)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftTxOnChain{}),
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
			WithNowNodes(&nowNodesTxNotFound{}),         // Not Found
		)
		provider, err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, provider)
	})

	t.Run("broadcast - tx already on-chain (WhatsOnChain)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMinercraft(&minerCraftTxNotFound{}), // Not found
			WithNowNodes(&nowNodesTxNotFound{}),     // Not Found
		)
		provider, err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, provider)
	})

	t.Run("broadcast - tx already on-chain (NowNodes)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodesTxOnChain{}),
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not found
		)
		provider, err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, provider)
	})
}

// TestClient_Broadcast_InMempool will test the method Broadcast()
func TestClient_Broadcast_InMempool(t *testing.T) {
	t.Parallel()

	t.Run("broadcast - in mempool (mAPI)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftInMempool{}),
			WithNowNodes(&nowNodesTxNotFound{}),         // Not Found
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
		)
		provider, err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, provider)
	})

	t.Run("broadcast - in mempool (WhatsOnChain)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithWhatsOnChain(&whatsOnChainInMempool{}),
			WithNowNodes(&nowNodesTxNotFound{}),     // Not Found
			WithMinercraft(&minerCraftTxNotFound{}), // Not found
		)
		provider, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, provider)
	})

	t.Run("broadcast - in mempool (NowNodes)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodeInMempool{}),
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not Found
		)
		provider, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
		assert.NotEmpty(t, provider)
	})
}

// TestClient_Broadcast will test the method Broadcast()
func TestClient_Broadcast(t *testing.T) {
	t.Parallel()

	t.Run("error - missing tx id", func(t *testing.T) {
		c := NewTestClient(context.Background(), t)
		provider, err := c.Broadcast(
			context.Background(), "", onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTransactionID)
		assert.Empty(t, provider)
	})

	t.Run("error - missing tx hex", func(t *testing.T) {
		c := NewTestClient(context.Background(), t)
		provider, err := c.Broadcast(
			context.Background(), onChainExample1TxID, "", defaultBroadcastTimeOut,
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTransactionHex)
		assert.Empty(t, provider)
	})

	t.Run("broadcast - all providers fail", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodesTxNotFound{}),         // Not found
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not Found
		)
		provider, err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.Error(t, err)
		assert.Equal(t, ProviderAll, provider)
	})
}

func containsAtLeastOneElement(coll1 []string, coll2 ...string) bool {
	m := make(map[string]bool)

	for _, element := range coll1 {
		m[element] = true
	}

	// Check if any element from bool  is present in the set
	for _, element := range coll2 {
		if m[element] {
			return true
		}
	}

	return false
}
