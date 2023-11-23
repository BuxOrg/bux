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
	version      uint32
	bumps        BUMPs
	transactions []*bt.Tx
}

func newBeefTx(ctx context.Context, version uint32, tx *Transaction) (*beefTx, error) {
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
		version:      version,
		bumps:        tx.draftTransaction.BUMPs,
		transactions: kahnTopologicalSortTransactions(transactions),
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

func saveBeefTransactionInput(ctx context.Context, c ClientInterface, input *bt.Tx) error {
	inputTx, err := c.GetTransactionByID(ctx, input.TxID())
	if err != nil && err != ErrMissingTransaction {
		return fmt.Errorf("error in saveBeefTransactionInput during getting transaction: %s", err.Error())
	}

	if inputTx != nil {
		if inputTx.BUMP.BlockHeight > 0 {
			return nil
		}

		// Sync tx if BUMP is empty
		err = _syncTxDataFromChain(ctx, inputTx.syncTransaction, inputTx)
		if err != nil {
			return fmt.Errorf("error in saveBeefTransactionInput during syncing transaction: %s", err.Error())
		}
		return nil
	}

	newOpts := c.DefaultModelOptions(New())
	inputTx = newTransaction(input.String(), newOpts...)

	err = inputTx.Save(ctx)
	if err != nil {
		return fmt.Errorf("error in saveBeefTransactionInput during saving tx: %s", err.Error())
	}

	sync := newSyncTransaction(
		inputTx.GetID(),
		inputTx.Client().DefaultSyncConfig(),
		inputTx.GetOptions(true)...,
	)
	sync.BroadcastStatus = SyncStatusSkipped
	sync.P2PStatus = SyncStatusSkipped
	sync.SyncStatus = SyncStatusReady

	if err = sync.Save(ctx); err != nil {
		return fmt.Errorf("error in saveBeefTransactionInput during saving sync tx: %s", err.Error())
	}

	err = _syncTxDataFromChain(ctx, sync, inputTx)
	if err != nil {
		return fmt.Errorf("error in saveBeefTransactionInput during syncing transaction: %s", err.Error())
	}
	return nil
}
