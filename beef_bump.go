package bux

import (
	"context"
	"errors"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

func calculateMergedBUMP(txs []*Transaction) (BUMPs, error) {
	bumps := make(map[uint64][]BUMP)
	mergedBUMPs := make(BUMPs, 0)

	for _, tx := range txs {
		if tx.BUMP.BlockHeight == 0 || len(tx.BUMP.Path) == 0 {
			return nil, fmt.Errorf("BUMP is not valid (tx.ID: %s)", tx.ID)
		}

		bumps[tx.BlockHeight] = append(bumps[tx.BlockHeight], tx.BUMP)
	}
	for _, b := range bumps {
		bump, err := CalculateMergedBUMP(b)
		if err != nil {
			return nil, fmt.Errorf("Error while calculating Merged BUMP: %s", err.Error())
		}
		if bump == nil {
			continue
		}
		mergedBUMPs = append(mergedBUMPs, bump)
	}

	return mergedBUMPs, nil
}

func validateBumps(bumps BUMPs) error {
	if len(bumps) == 0 {
		return errors.New("empty bump paths slice")
	}

	for _, p := range bumps {
		if len(p.Path) == 0 {
			return errors.New("one of bump path is empty")
		}
	}

	return nil
}

// getClientTransactions is a helper function to get transactions in batch to reduce API calls.
func getClientTransactions(ctx context.Context, client ClientInterface, txIds []string) (map[string]*bt.Tx, error) {
	transactions := make(map[string]*bt.Tx)
	for _, txId := range txIds {
		tx, err := client.GetTransactionByID(ctx, txId)
		if err != nil {
			return nil, err
		}
		btTx, err := bt.NewTxFromString(tx.Hex)
		if err != nil {
			return nil, err
		}
		transactions[txId] = btTx
	}
	return transactions, nil
}

// prepareBUMPFactors is a recursive function to find all parent transactions
// with a valid MerkleProof or BUMP.
func prepareBUMPFactors(ctx context.Context, client ClientInterface, inputs []*TransactionInput) ([]*bt.Tx, []*Transaction, error) {
	var btTxsNeededForBUMP []*bt.Tx
	var txsNeededForBUMP []*Transaction

	for _, input := range inputs {
		// TODO: Before finishing I will try to move to client.GetTransactions to reduce calls. Need to dig into the metadata construction.
		inputTx, err := client.GetTransactionByID(ctx, input.UtxoPointer.TransactionID)
		if err != nil {
			return nil, nil, err
		}

		inputBtTx, err := bt.NewTxFromString(inputTx.Hex)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot convert to bt.Tx from hex (tx.ID: %s). Reason: %w", inputTx.ID, err)
		}

		if inputTx.MerkleProof.TxOrID != "" || inputTx.BUMP.BlockHeight != 0 || len(inputTx.BUMP.Path) != 0 {
			txsNeededForBUMP = append(txsNeededForBUMP, inputTx)
			btTxsNeededForBUMP = append(btTxsNeededForBUMP, inputBtTx)
		} else {
			parentBtTransactions, parentTransactions, err := checkParentTransactions(ctx, client, inputTx)
			if err != nil {
				return nil, nil, err
			}

			txsNeededForBUMP = append(txsNeededForBUMP, parentTransactions...)
			btTxsNeededForBUMP = append(btTxsNeededForBUMP, parentBtTransactions...)
		}
	}

	return btTxsNeededForBUMP, txsNeededForBUMP, nil
}

// checkParentTransactions is a helper recursive function to check the parent transactions.
func checkParentTransactions(ctx context.Context, client ClientInterface, inputTx *Transaction) ([]*bt.Tx, []*Transaction, error) {
	btTx, err := bt.NewTxFromString(inputTx.Hex)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot convert to bt.Tx from hex (tx.ID: %s). Reason: %w", inputTx.ID, err)
	}

	var validTxs []*Transaction
	var validBtTxs []*bt.Tx
	for _, txIn := range btTx.Inputs {
		parentTx, err := client.GetTransactionByID(ctx, txIn.PreviousTxIDStr())
		if err != nil {
			return nil, nil, err
		}

		// If the parent transaction has a MerkleProof or a BUMP, add it to the list.
		if parentTx.MerkleProof.TxOrID != "" || parentTx.BUMP.BlockHeight != 0 {
			parentBtTx, err := bt.NewTxFromString(parentTx.Hex)
			if err != nil {
				return nil, nil, err
			}
			validTxs = append(validTxs, parentTx)
			validBtTxs = append(validBtTxs, parentBtTx)
		} else {
			// Otherwise, recursively check the parents of this parent.
			parentValidBtTxs, parentValidTxs, err := checkParentTransactions(ctx, client, parentTx)
			if err != nil {
				return nil, nil, err
			}
			validTxs = append(validTxs, parentValidTxs...)
			validBtTxs = append(validBtTxs, parentValidBtTxs...)
		}
	}

	if len(validBtTxs) == 0 {
		return nil, nil, fmt.Errorf("transaction is not mined yet (tx.ID: %s)", inputTx.ID)
	}

	return validBtTxs, validTxs, nil
}
