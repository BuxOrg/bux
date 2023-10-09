package bux

import (
	"fmt"
	"math/rand"
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

	txsFromOldestToNewestWithUnnecessaryInputs := []*Transaction{
		createTx("0"),
		createTx("1", "0"),
		createTx("2", "1", "101", "102"),
		createTx("3", "2", "1"),
		createTx("4", "3", "1"),
		createTx("5", "3", "2", "100"),
		createTx("6", "4", "2", "0"),
		createTx("7", "6", "5", "3", "1", "103", "105", "106"),
		createTx("8", "7"),
	}

	tCases := []struct {
		name                       string
		expectedSortedTransactions []*Transaction
	}{{
		name:                       "txs with necessary data only",
		expectedSortedTransactions: txsFromOldestToNewest,
	},
		{
			name:                       "txs with inputs from other txs",
			expectedSortedTransactions: txsFromOldestToNewestWithUnnecessaryInputs,
		},
	}

	for _, tc := range tCases {
		t.Run(fmt.Sprint("sort from oldest to newest ", tc.name), func(t *testing.T) {
			// given
			unsortedTxs := shuffleTransactions(tc.expectedSortedTransactions)

			// when
			sortedGraph := kahnTopologicalSortTransactions(unsortedTxs)

			// then
			for i, tx := range txsFromOldestToNewest {
				assert.Equal(t, tx.ID, sortedGraph[i].ID)
			}
		})
	}
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

func shuffleTransactions(txs []*Transaction) []*Transaction {
	n := len(txs)
	result := make([]*Transaction, n)
	copy(result, txs)

	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result
}
