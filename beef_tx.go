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

func ToBeef(ctx context.Context, tx *Transaction) (string, error) {
	if err := hydrateTransaction(ctx, tx); err != nil {
		return "", err
	}

	bumpBtFactors, bumpFactors, err := prepareBUMPFactors(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("prepareBUMPFactors() error: %w", err)
	}

	tx.draftTransaction.BUMPs, err = calculateMergedBUMP(bumpFactors)
	sortedTxs := kahnTopologicalSortTransactions(bumpBtFactors)
	beefHex, err := toBeefHex(ctx, tx, sortedTxs)
	if err != nil {
		return "", fmt.Errorf("ToBeef() error: %w", err)
	}

	return beefHex, nil
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

func newBeefTx(ctx context.Context, version uint32, tx *Transaction, parentTxs []*bt.Tx) (*beefTx, error) {
	if version > maxBeefVer {
		return nil, fmt.Errorf("version above 0x%X", maxBeefVer)
	}

	if err := validateBumps(tx.draftTransaction.BUMPs); err != nil {
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
