package bux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_kahnTopologicalSortTransaction(t *testing.T) {
	// create related transactions from oldest to newest
	txsFromOldestToNewest := []*Transaction{
		createTx("0"),
		createTx("1", "0"),
		createTx("2", "1"),
		createTx("3", "2", "1"),
		createTx("4", "3", "1"),
		createTx("5", "3", "2"),
		createTx("6", "4", "2", "0"),
		createTx("7", "6", "5", "3", "1"),
		createTx("8", "7"),
	}

	unsortedTxs := []*Transaction{
		txsFromOldestToNewest[2],
		txsFromOldestToNewest[3],
		txsFromOldestToNewest[0],
		txsFromOldestToNewest[1],
		txsFromOldestToNewest[4],
		txsFromOldestToNewest[7],
		txsFromOldestToNewest[5],
		txsFromOldestToNewest[6],
		txsFromOldestToNewest[8],
	}

	t.Run("kahnTopologicalSortTransaction sort from oldest to newest", func(t *testing.T) {
		sortedGraph := kahnTopologicalSortTransactions(unsortedTxs)

		for i, tx := range txsFromOldestToNewest {
			assert.Equal(t, tx.ID, sortedGraph[i].ID)
		}

	})
}

func createTx(txID string, inputsTxIDs ...string) *Transaction {
	inputs := make([]*TransactionInput, 0)
	for _, inTxID := range inputsTxIDs {
		in := &TransactionInput{
			Utxo: Utxo{
				UtxoPointer: UtxoPointer{
					TransactionID: inTxID,
				},
			},
		}

		inputs = append(inputs, in)
	}

	transaction := &Transaction{
		draftTransaction: &DraftTransaction{
			Configuration: TransactionConfig{
				Inputs: inputs,
			},
		},
	}

	transaction.ID = txID

	return transaction
}
