package bux

import (
	"context"
	"testing"

	"github.com/libsv/go-bc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ToBeefHex(t *testing.T) {
	// TOOD: prepare tests in BUX-168

	t.Run("some parents txs are not mined yet", func(t *testing.T) {
		// Error expected! this should be changed in the future. right now the test case has been written to make sure the system doesn't panic in such a situation

		// given
		ctx, client, deferMe := initSimpleTestCase(t)
		defer deferMe()

		ancestorTx := addGrandpaTx(ctx, t, client)
		notMinedParentTx := createTxWithDraft(ctx, t, client, ancestorTx, false)

		newTx := createTxWithDraft(ctx, t, client, notMinedParentTx, false)

		// when
		hex, err := ToBeefHex(ctx, newTx)

		// then
		assert.Error(t, err)
		assert.Empty(t, hex)
	})
}

func addGrandpaTx(ctx context.Context, t *testing.T, client ClientInterface) *Transaction {
	// great ancestor
	grandpaTx := newTransaction(testTx2Hex, append(client.DefaultModelOptions(), New())...)
	grandpaTx.BlockHeight = 1
	// mark it as mined
	grandpaTxMp := bc.MerkleProof{
		TxOrID: "cefffc5415620292081f7e941bb74d11a3188144312c4d7550c462b2a151c64d",
		Nodes: []string{
			"6cf512411d03ab9b61643515e7aa9afd005bf29e1052ade95410b3475f02820c",
			"cd73c0c6bb645581816fa960fd2f1636062fcbf23cb57981074ab8d708a76e3b",
			"b4c8d919190a090e77b73ffcd52b85babaaeeb62da000473102aca7f070facef",
			"3470d882cf556a4b943639eba15dc795dffdbebdc98b9a98e3637fda96e3811e",
		},
	}
	grandpaTx.MerkleProof = MerkleProof(grandpaTxMp)
	grandpaTx.BUMP = grandpaTx.MerkleProof.ToBUMP(grandpaTx.BlockHeight)
	err := grandpaTx.Save(ctx)
	require.NoError(t, err)

	return grandpaTx
}

func createTxWithDraft(ctx context.Context, t *testing.T, client ClientInterface, parentTx *Transaction, mined bool) *Transaction {
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
