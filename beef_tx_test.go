package bux

import (
	"context"
	"testing"

	"github.com/libsv/go-bc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ToBeefHex(t *testing.T) {
	t.Run("all parents txs are already mined", func(t *testing.T) {
		//given
		ctx, client, deferMe := initSimpleTestCase(t)
		defer deferMe()

		ancestorTx := addGrandPaTx(t, ctx, client)
		minedParentTx := createTxWithDraft(t, ctx, client, ancestorTx, true)

		newTx := createTxWithDraft(t, ctx, client, minedParentTx, false)

		//when
		hex, err := ToBeefHex(ctx, newTx)

		//then
		assert.NoError(t, err)
		assert.NotEmpty(t, hex)
	})

	t.Run("some parents txs are not mined yet", func(t *testing.T) {
		// Error expeted! this should be changed in the future. right now the test case has been written to make sure the system doesn't panic in such a situation

		//given
		ctx, client, deferMe := initSimpleTestCase(t)
		defer deferMe()

		ancestorTx := addGrandPaTx(t, ctx, client)
		notMinedParentTx := createTxWithDraft(t, ctx, client, ancestorTx, false)

		newTx := createTxWithDraft(t, ctx, client, notMinedParentTx, false)

		//when
		hex, err := ToBeefHex(ctx, newTx)

		//then
		assert.Error(t, err)
		assert.Empty(t, hex)
	})
}

func addGrandPaTx(t *testing.T, ctx context.Context, client ClientInterface) *Transaction {
	// great ancestor
	grandpaTx := newTransaction(testTx2Hex, append(client.DefaultModelOptions(), New())...)
	grandpaTx.BlockHeight = 1
	// mark it as mined
	grandpaTxMp := bc.MerkleProof{
		TxOrID: "111111111111111111111111111111111111111",
		Nodes:  []string{"n1", "n2"},
	}
	grandpaTx.MerkleProof = MerkleProof(grandpaTxMp)
	err := grandpaTx.Save(ctx)
	require.NoError(t, err)

	return grandpaTx
}

func createTxWithDraft(t *testing.T, ctx context.Context, client ClientInterface, parentTx *Transaction, mined bool) *Transaction {
	draftTransaction := newDraftTransaction(
		testXPub, &TransactionConfig{
			Inputs: []*TransactionInput{{Utxo: *newUtxoFromTxID(parentTx.GetID(), 0, append(client.DefaultModelOptions(), New())...)}},
			Outputs: []*TransactionOutput{{
				To:       "1A1PjKqjWMNBzTVdcBru27EV1PHcXWc63W",
				Satoshis: 1000,
			}},
			ChangeNumberOfDestinations: 1,
			Sync: &SyncConfig{
				Broadcast:        true,
				BroadcastInstant: false,
				PaymailP2P:       false,
				SyncOnChain:      false,
			},
		},
		append(client.DefaultModelOptions(), New())...,
	)

	err := draftTransaction.Save(ctx)
	require.NoError(t, err)

	var transaction *Transaction
	transaction, err = client.RecordTransaction(ctx, testXPub, draftTransaction.Hex, draftTransaction.ID, client.DefaultModelOptions()...)
	require.NoError(t, err)
	assert.NotEmpty(t, transaction)

	if mined {
		transaction.BlockHeight = 128
		mp := bc.MerkleProof{
			TxOrID: "423542156234627frafserg6gtrdsbd", Nodes: []string{"n1", "n2"},
		}
		transaction.MerkleProof = MerkleProof(mp)
	}

	err = transaction.Save(ctx)
	require.NoError(t, err)

	return transaction
}
