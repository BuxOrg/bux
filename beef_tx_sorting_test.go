package bux

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

func Test_kahnTopologicalSortTransaction(t *testing.T) {

	tCases := []struct {
		name                       string
		expectedSortedTransactions []*bt.Tx
	}{
		{
			name:                       "txs with necessary data only",
			expectedSortedTransactions: getTxsFromOldestToNewestWithNecessaryDataOnly(),
		},
		{
			name:                       "txs with inputs from other txs",
			expectedSortedTransactions: getTxsFromOldestToNewestWithUnecessaryData(),
		},
	}

	for _, tc := range tCases {
		t.Run(fmt.Sprint("sort from oldest to newest ", tc.name), func(t *testing.T) {
			// given
			unsortedTxs := shuffleTransactions(tc.expectedSortedTransactions)

			// when
			sortedGraph := kahnTopologicalSortTransactions(unsortedTxs)

			// then
			for i, tx := range tc.expectedSortedTransactions {
				assert.Equal(t, tx.TxID(), sortedGraph[i].TxID())
			}
		})
	}
}

func getTxsFromOldestToNewestWithNecessaryDataOnly() []*bt.Tx {
	// create related transactions from oldest to newest
	oldestTx := createTx()
	secondTx := createTx(oldestTx)
	thirdTx := createTx(secondTx)
	fourthTx := createTx(thirdTx, secondTx)
	fifthTx := createTx(fourthTx, secondTx)
	sixthTx := createTx(fourthTx, thirdTx)
	seventhTx := createTx(fifthTx, thirdTx, oldestTx)
	eightTx := createTx(seventhTx, sixthTx, fourthTx, secondTx)

	newestTx := createTx(eightTx)

	txsFromOldestToNewest := []*bt.Tx{
		oldestTx,
		secondTx,
		thirdTx,
		fourthTx,
		fifthTx,
		sixthTx,
		seventhTx,
		eightTx,
		newestTx,
	}

	return txsFromOldestToNewest
}

func getTxsFromOldestToNewestWithUnecessaryData() []*bt.Tx {
	unnecessaryParentTx_1 := createTx()
	unnecessaryParentTx_2 := createTx()
	unnecessaryParentTx_3 := createTx()
	unnecessaryParentTx_4 := createTx()

	// create related transactions from oldest to newest
	oldestTx := createTx()
	secondTx := createTx(oldestTx)
	thirdTx := createTx(secondTx)
	fourthTx := createTx(thirdTx, secondTx, unnecessaryParentTx_1, unnecessaryParentTx_4)
	fifthTx := createTx(fourthTx, secondTx)
	sixthTx := createTx(fourthTx, thirdTx, unnecessaryParentTx_3, unnecessaryParentTx_2, unnecessaryParentTx_1)
	seventhTx := createTx(fifthTx, thirdTx, oldestTx)
	eightTx := createTx(seventhTx, sixthTx, fourthTx, secondTx, unnecessaryParentTx_1)

	newestTx := createTx(eightTx)

	txsFromOldestToNewest := []*bt.Tx{
		oldestTx,
		secondTx,
		thirdTx,
		fourthTx,
		fifthTx,
		sixthTx,
		seventhTx,
		eightTx,
		newestTx,
	}

	return txsFromOldestToNewest
}

func createTx(inputsParents ...*bt.Tx) *bt.Tx {
	inputs := make([]*bt.Input, 0)

	for _, parent := range inputsParents {
		in := bt.Input{}
		in.PreviousTxIDAdd(parent.TxIDBytes())

		inputs = append(inputs, &in)
	}

	transaction := bt.NewTx()
	transaction.Inputs = append(transaction.Inputs, inputs...)

	return transaction
}

func shuffleTransactions(txs []*bt.Tx) []*bt.Tx {
	n := len(txs)
	result := make([]*bt.Tx, n)
	copy(result, txs)

	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result
}
