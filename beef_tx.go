package bux

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

type ToConstructBUMP struct {
	BUMPInputs []*BUMPInputs
}

type BUMPInputs struct {
	Input              *bt.Tx
	HasBUMP            bool
	ParentTransactions []*bt.Tx
}

const maxBeefVer = uint32(0xFFFF) // value from BRC-62

func ToBeef(ctx context.Context, tx *Transaction) (string, error) {
	// get inputs parent transactions
	// inputs := tx.draftTransaction.Configuration.Inputs
	// transactions := make([]*bt.Tx, 0, len(inputs)+1)

	// for _, input := range inputs {
	// 	var prevTxs []*bt.Tx
	// 	prevTxs, err = getParentTransactionsForInput(ctx, tx.client, input)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("retrieve input parent transaction failed: %w", err)
	// 	}

	// 	transactions = append(transactions, prevTxs...)
	// }

	// // add current transaction
	// var btTx *bt.Tx
	// btTx, err = bt.NewTxFromString(tx.Hex)
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot convert new transaction to bt.Tx from hex (tx.ID: %s). Reason: %w", tx.ID, err)
	// }
	// transactions = append(transactions, btTx)

	// khanTopologicalSortTransactions(transactions)
	return "", nil
}

// toBeefHex generates BEEF Hex for transaction
func toBeefHex(ctx context.Context, tx *Transaction, parentTxs []*bt.Tx) (string, error) {
	beef, err := newBeefTx(ctx, 1, tx, parentTxs)
	if err != nil {
		return "", fmt.Errorf("ToBeefHex() error: %w", err)
	}

	beefBytes, err := beef.toBeefBytes()
	if err != nil {
		return "", fmt.Errorf("ToBeefHex() error: %w", err)
	}

	return hex.EncodeToString(beefBytes), nil
}

type beefTx struct {
	version      uint32
	bumps        BUMPs
	transactions []*bt.Tx
}

func newBeefTx(ctx context.Context, version uint32, tx *Transaction, parentTxs []*bt.Tx) (*beefTx, error) {
	if version > maxBeefVer {
		return nil, fmt.Errorf("version above 0x%X", maxBeefVer)
	}

	var err error
	if err = hydrateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	if err = validateBumps(tx.draftTransaction.BUMPs); err != nil {
		return nil, err
	}

	beef := &beefTx{
		version:      version,
		bumps:        tx.draftTransaction.BUMPs,
		transactions: parentTxs,
	}

	return beef, nil
}

func hydrateTransaction(ctx context.Context, tx *Transaction) error {
	if tx.draftTransaction == nil {
		dTx, err := getDraftTransactionID(
			ctx, tx.XPubID, tx.DraftID, tx.GetOptions(false)...,
		)

		if err != nil || dTx == nil {
			return fmt.Errorf("retrieve DraftTransaction failed: %w", err)
		}

		tx.draftTransaction = dTx
	}

	return nil
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

// func getParentTransactionsForInput(ctx context.Context, client ClientInterface, input *TransactionInput) ([]*bt.Tx, error) {
// 	inputTx, err := client.GetTransactionByID(ctx, input.UtxoPointer.TransactionID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	inputBtTx, err := bt.NewTxFromString(inputTx.Hex)

// 	if inputTx.MerkleProof.TxOrID != "" || inputTx.BUMP.BlockHeight != 0 {
// 		if err != nil {
// 			return nil, fmt.Errorf("cannot convert to bt.Tx from hex (tx.ID: %s). Reason: %w", inputTx.ID, err)
// 		}

// 		return []*bt.Tx{inputBtTx}, nil
// 	} else {
// 		parentInputs := inputBtTx.Inputs
// 		parentInputsIds := make([]string, 0, len(parentInputs))
// 		for _, parentInput := range parentInputs {
// 			parentInputsIds = append(parentInputsIds, parentInput.PreviousTxIDStr())
// 		}

// 		parentInputs := client.GetTransactions(ctx, parentInputsIds)

// 		// sprawdź czy któryś z parentów nie spełnia tego warunku:
// 		// if inputTx.MerkleProof.TxOrID != "" || inputTx.BUMP.BlockHeight != 0

// 		// jeśli tak to weź jego parentów i postępuj tak do momentu az wszystkie parenty na danym poziomie beda spelniac ten warunek
// 		// zwroc wszystkie transakcje ktore maja bumpa i merkle proof

// 		return nil, nil
// 	}

// 	return nil, fmt.Errorf("transaction is not mined yet (tx.ID: %s)", inputTx.ID) // TODO: handle it in next iterration
// }

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

// getParentTransactionsForInput is a recursive function to find all parent transactions
// with a valid MerkleProof or BUMP.
func getParentTransactionsForInput(ctx context.Context, client ClientInterface, inputs []*TransactionInput) (*ToConstructBUMP, error) {
	var toConstructBUMP ToConstructBUMP

	for _, input := range inputs {
		inputTx, err := client.GetTransactionByID(ctx, input.UtxoPointer.TransactionID)
		if err != nil {
			return nil, err
		}

		inputBtTx, err := bt.NewTxFromString(inputTx.Hex)
		if err != nil {
			return nil, fmt.Errorf("cannot convert to bt.Tx from hex (tx.ID: %s). Reason: %w", inputTx.ID, err)
		}

		if inputTx.MerkleProof.TxOrID != "" || inputTx.BUMP.BlockHeight != 0 {
			toConstructBUMP.BUMPInputs = append(toConstructBUMP.BUMPInputs, &BUMPInputs{
				Input:   inputBtTx,
				HasBUMP: true,
			})
		} else {
			parentTransactions, err := checkParentTransactions(ctx, client, inputTx)
			if err != nil {
				return nil, err
			}
			toConstructBUMP.BUMPInputs = append(toConstructBUMP.BUMPInputs, &BUMPInputs{
				Input:              inputBtTx,
				HasBUMP:            false,
				ParentTransactions: parentTransactions,
			})
		}
	}

	return &toConstructBUMP, nil
}

// checkParentTransactions is a helper recursive function to check the parent transactions.
func checkParentTransactions(ctx context.Context, client ClientInterface, inputTx *Transaction) ([]*bt.Tx, error) {
	btTx, err := bt.NewTxFromString(inputTx.Hex)
	if err != nil {
		return nil, fmt.Errorf("cannot convert to bt.Tx from hex (tx.ID: %s). Reason: %w", inputTx.ID, err)
	}

	var validTxs []*bt.Tx
	for _, txIn := range btTx.Inputs {
		parentTx, err := client.GetTransactionByID(ctx, txIn.PreviousTxIDStr())
		if err != nil {
			return nil, err
		}

		// If the parent transaction has a MerkleProof or a BUMP, add it to the list.
		if parentTx.MerkleProof.TxOrID != "" || parentTx.BUMP.BlockHeight != 0 {
			parentBtTx, err := bt.NewTxFromString(parentTx.Hex)
			if err != nil {
				return nil, err
			}
			validTxs = append(validTxs, parentBtTx)
		} else {
			// Otherwise, recursively check the parents of this parent.
			parentValidTxs, err := checkParentTransactions(ctx, client, parentTx)
			if err != nil {
				return nil, err
			}
			validTxs = append(validTxs, parentValidTxs...)
		}
	}

	if len(validTxs) == 0 {
		return nil, fmt.Errorf("transaction is not mined yet (tx.ID: %s)", inputTx.ID)
	}

	return validTxs, nil
}
