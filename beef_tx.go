package bux

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

const maxBeefVer = uint32(0xFFFF) // value from BRC-62

// ToBeefHex generates BEEF Hex for transaction
func ToBeefHex(ctx context.Context, tx *Transaction) (string, error) {
	beef, err := newBeefTx(ctx, 1, tx)
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
	version             uint32
	compoundMerklePaths CMPSlice
	transactions        []*bt.Tx
}

func newBeefTx(ctx context.Context, version uint32, tx *Transaction) (*beefTx, error) {
	if version > maxBeefVer {
		return nil, fmt.Errorf("version above 0x%X", maxBeefVer)
	}

	var err error
	if err = hydrateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	if err = validateCompoundMerklePathes(tx.draftTransaction.CompoundMerklePathes); err != nil {
		return nil, err
	}

	// get inputs parent transactions
	inputs := tx.draftTransaction.Configuration.Inputs
	transactions := make([]*bt.Tx, 0, len(inputs)+1)

	for _, input := range inputs {
		var prevTxs []*bt.Tx
		prevTxs, err = getParentTransactionsForInput(ctx, tx.client, input)
		if err != nil {
			return nil, fmt.Errorf("retrieve input parent transaction failed: %w", err)
		}

		transactions = append(transactions, prevTxs...)
	}

	// add current transaction
	var btTx *bt.Tx
	btTx, err = bt.NewTxFromString(tx.Hex)
	if err != nil {
		return nil, fmt.Errorf("cannot convert new transaction to bt.Tx from hex (tx.ID: %s). Reason: %w", tx.ID, err)
	}
	transactions = append(transactions, btTx)

	beef := &beefTx{
		version:             version,
		compoundMerklePaths: tx.draftTransaction.CompoundMerklePathes,
		transactions:        kahnTopologicalSortTransactions(transactions),
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

func validateCompoundMerklePathes(compountedPaths CMPSlice) error {
	if len(compountedPaths) == 0 {
		return errors.New("empty compounted paths slice")
	}

	for _, c := range compountedPaths {
		if len(c) == 0 {
			return errors.New("one of compounted merkle paths is empty")
		}
	}

	return nil
}

func getParentTransactionsForInput(ctx context.Context, client ClientInterface, input *TransactionInput) ([]*bt.Tx, error) {
	inputTx, err := client.GetTransactionByID(ctx, input.UtxoPointer.TransactionID)
	if err != nil {
		return nil, err
	}

	if inputTx.MerkleProof.TxOrID != "" {
		inputBtTx, err := bt.NewTxFromString(inputTx.Hex)
		if err != nil {
			return nil, fmt.Errorf("cannot convert to bt.Tx from hex (tx.ID: %s). Reason: %w", inputTx.ID, err)
		}

		return []*bt.Tx{inputBtTx}, nil
	}

	return nil, fmt.Errorf("transaction is not mined yet (tx.ID: %s)", inputTx.ID) // TODO: handle it in next iterration
}
