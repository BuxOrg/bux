package chainstate

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			WithMatterCloud(&matterCloudTxNotFound{}),   // Not Found
		)
		err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - success (MatterCloud)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMatterCloud(&matterCloudBroadcastSuccess{}),
			WithNowNodes(&nowNodesTxNotFound{}),         // Not Found
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not found
		)
		err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - success (WhatsOnChain)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithWhatsOnChain(&whatsOnChainBroadcastSuccess{}),
			WithMatterCloud(&matterCloudTxNotFound{}), // Not Found
			WithNowNodes(&nowNodesTxNotFound{}),       // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),   // Not found
		)
		err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - success (NowNodes)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodesBroadcastSuccess{}),
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
			WithMatterCloud(&matterCloudTxNotFound{}),   // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not found
		)
		err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
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
			WithMatterCloud(&matterCloudTxNotFound{}),   // Not Found
			WithNowNodes(&nowNodesTxNotFound{}),         // Not Found
		)
		err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - tx already on-chain (MatterCloud)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMatterCloud(&matterCloudTxOnChain{}),
			WithMinercraft(&minerCraftTxNotFound{}),     // Not found
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
			WithNowNodes(&nowNodesTxNotFound{}),         // Not Found
		)
		err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - tx already on-chain (WhatsOnChain)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithWhatsOnChain(&whatsOnChainTxOnChain{}),
			WithMatterCloud(&matterCloudTxNotFound{}), // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),   // Not found
			WithNowNodes(&nowNodesTxNotFound{}),       // Not Found
		)
		err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - tx already on-chain (NowNodes)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodesTxOnChain{}),
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
			WithMatterCloud(&matterCloudTxNotFound{}),   // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not found
		)
		err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})
}

// TestClient_Broadcast_InMempool will test the method Broadcast()
func TestClient_Broadcast_InMempool(t *testing.T) {
	t.Parallel()

	t.Run("broadcast - in mempool (mAPI)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMinercraft(&minerCraftInMempool{}),
			WithMatterCloud(&matterCloudTxNotFound{}),   // Not found
			WithNowNodes(&nowNodesTxNotFound{}),         // Not Found
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
		)
		err := c.Broadcast(
			context.Background(), onChainExample1TxID, onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - in mempool (MatterCloud)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithMatterCloud(&matterCloudInMempool{}),
			WithNowNodes(&nowNodesTxNotFound{}),         // Not Found
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not found
		)
		err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - in mempool (WhatsOnChain)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithWhatsOnChain(&whatsOnChainInMempool{}),
			WithMatterCloud(&matterCloudTxNotFound{}), // Not Found
			WithNowNodes(&nowNodesTxNotFound{}),       // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),   // Not found
		)
		err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})

	t.Run("broadcast - in mempool (NowNodes)", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodeInMempool{}),
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not found
			WithMatterCloud(&matterCloudTxNotFound{}),   // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not Found
		)
		err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.NoError(t, err)
	})
}

// TestClient_Broadcast will test the method Broadcast()
func TestClient_Broadcast(t *testing.T) {
	t.Parallel()

	t.Run("error - missing tx id", func(t *testing.T) {
		c := NewTestClient(context.Background(), t)
		err := c.Broadcast(
			context.Background(), "", onChainExample1TxHex, defaultBroadcastTimeOut,
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTransactionID)
	})

	t.Run("error - missing tx hex", func(t *testing.T) {
		c := NewTestClient(context.Background(), t)
		err := c.Broadcast(
			context.Background(), onChainExample1TxID, "", defaultBroadcastTimeOut,
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidTransactionHex)
	})

	t.Run("broadcast - all providers fail", func(t *testing.T) {
		c := NewTestClient(
			context.Background(), t,
			WithNowNodes(&nowNodesTxNotFound{}),         // Not found
			WithWhatsOnChain(&whatsOnChainTxNotFound{}), // Not found
			WithMatterCloud(&matterCloudTxNotFound{}),   // Not Found
			WithMinercraft(&minerCraftTxNotFound{}),     // Not Found
		)
		err := c.Broadcast(
			context.Background(), broadcastExample1TxID, broadcastExample1TxHex, defaultBroadcastTimeOut,
		)
		require.Error(t, err)
	})

	t.Run("broadcast - bad tx - error mempool conflict?", func(t *testing.T) {

	})

	t.Run("broadcast - server errors", func(t *testing.T) {

	})
}
