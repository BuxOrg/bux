package bux

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

const maxBeefVer = uint32(0xFFFF) // value from BRC-62

type beefTx struct {
	version      uint32
	bumps        BUMPs
	transactions []*bt.Tx
}

// ToBeef generates BEEF Hex for transaction
func ToBeef(ctx context.Context, tx *Transaction, store TransactionGetter) (string, error) {
	if err := hydrateTransaction(ctx, tx); err != nil {
		return "", err
	}

	bumpBtFactors, bumpFactors, err := prepareBEEFFactors(ctx, tx, store)
	if err != nil {
		return "", fmt.Errorf("prepareBUMPFactors() error: %w", err)
	}

	bumps, err := calculateMergedBUMP(bumpFactors)
	sortedTxs := kahnTopologicalSortTransactions(bumpBtFactors)
	beefHex, err := toBeefHex(ctx, bumps, sortedTxs)
	if err != nil {
		return "", fmt.Errorf("ToBeef() error: %w", err)
	}

	return beefHex, nil
}

func toBeefHex(ctx context.Context, bumps BUMPs, parentTxs []*bt.Tx) (string, error) {
	beef, err := newBeefTx(ctx, 1, bumps, parentTxs)
	if err != nil {
		return "", fmt.Errorf("ToBeefHex() error: %w", err)
	}

	beefBytes, err := beef.toBeefBytes()
	if err != nil {
		return "", fmt.Errorf("ToBeefHex() error: %w", err)
	}

	return hex.EncodeToString(beefBytes), nil
}

func newBeefTx(ctx context.Context, version uint32, bumps BUMPs, parentTxs []*bt.Tx) (*beefTx, error) {
	if version > maxBeefVer {
		return nil, fmt.Errorf("version above 0x%X", maxBeefVer)
	}

	if err := validateBumps(bumps); err != nil {
		return nil, err
	}

	beef := &beefTx{
		version:      version,
		bumps:        bumps,
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
			fmt.Println("input exist in db and has BUMP")
			return nil
		}
		fmt.Println("input exist in db and has no BUMP")

		// Sync tx if BUMP is empty
		err = _syncTxDataFromChain(ctx, inputTx.syncTransaction, inputTx)
		if err != nil {
			return fmt.Errorf("error in saveBeefTransactionInput during syncing transaction: %s", err.Error())
		}
		return nil
	}

	fmt.Println("input not exist in db")

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
