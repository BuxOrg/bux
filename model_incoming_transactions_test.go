package bux

import (
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIncomingTransaction_GetModelName will test the method GetModelName()
func TestIncomingTransaction_GetModelName(t *testing.T) {
	t.Parallel()

	bTx := newIncomingTransaction(testTxID, testTxHex, New())
	assert.Equal(t, ModelIncomingTransaction.String(), bTx.GetModelName())
}

// TestProcessIncomingTransaction will test the method processIncomingTransaction()
func (ts *EmbeddedDBTestSuite) TestProcessIncomingTransaction() {

	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - LIVE integration test - valid external incoming tx", func(t *testing.T) {

			// todo: mock the response vs using a LIVE request for Chainstate

			tc := ts.genericDBClient(t, testCase.database, true)
			defer tc.Close(tc.ctx)

			// Create a xpub
			var err error
			xPubKey := "xpub6826nizKsKjNvxGbcYPiyS4tLVB3nd3e4yujBe6YmqmNtN3DMytsQMkruEgHoyUu89CHcTtaeeLynTC19fD4JcAvKXBUbHi9qdeWtUMYCQK"
			xPub := newXpub(xPubKey, append(tc.client.DefaultModelOptions(), New())...)
			require.NotNil(t, xPub)

			err = xPub.Save(tc.ctx)
			require.NoError(t, err)

			// Create a destination
			var destination *Destination
			destination, err = xPub.getNewDestination(tc.ctx, utils.ChainExternal, utils.ScriptTypePubKeyHash, tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			require.NotNil(t, destination)

			// Save the updated xPub and new destination
			err = xPub.Save(tc.ctx)
			require.NoError(t, err)

			// Record an external incoming tx
			txHex := "0100000001574eacf3305f561f63d6f1896566d5ff63409fea2aae1534a3e3734191b47430020000006b483045022100e3f002e318d2dfae67f00da8aa327cc905e93d4a5adb5b7c33afde95bfc26acc022000ddfcdba500e0ba9eaadde478e2b6c6566f8d6837e7802c5f867492eadfe5d1412102ff596abfae0099d480d93937380af985f5165b84ad31790c10c09d3daab8562effffffff01493a1100000000001976a914ec8470c5d9275c39829b15ea7f1997cb66082d3188ac00000000"
			var tx *Transaction
			tx, err = tc.client.RecordTransaction(tc.ctx, xPubKey, txHex, "", tc.client.DefaultModelOptions()...)
			require.NoError(t, err)
			require.NotNil(t, tx)

			// Process if found
			err = processIncomingTransactions(tc.ctx, 5, WithClient(tc.client))
			require.NoError(t, err)

			// Check if the tx is found in the datastore
			var foundTx *Transaction
			foundTx, err = tc.client.GetTransaction(tc.ctx, xPub.ID, tx.ID)
			require.NoError(t, err)
			require.NotNil(t, foundTx)

			// Test that we found the tx on-chain
			assert.Equal(t, "0000000000000000090eedbbfef5c542e252fee42db4ac54fa59ca5dbad510b4", foundTx.BlockHash)
			assert.Equal(t, uint64(742742), foundTx.BlockHeight)
		})
	}
}
