package bux

import (
	"context"
	"fmt"
	"time"

	"github.com/libsv/go-bt"
)

type recordTxStrategy interface {
	TxID() string
	Validate() error
	Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error)
}

type recordIncomingTxStrategy interface {
	ForceBroadcast(force bool)
}

func recordTransaction(ctx context.Context, c ClientInterface, strategy recordTxStrategy, opts ...ModelOps) (*Transaction, error) {
	unlock := waitForRecordTxWriteLock(ctx, c, strategy.TxID())
	defer unlock()

	transaction, err := strategy.Execute(ctx, c, opts)
	return transaction, err
}

func getRecordTxStrategy(ctx context.Context, c ClientInterface, xPubKey, txHex, draftID string) (recordTxStrategy, error) {
	var rts recordTxStrategy

	if draftID != "" {
		rts = &outgoingTx{
			Hex:            txHex,
			RelatedDraftID: draftID,
			XPubKey:        xPubKey,
		}
	} else {
		tx, err := getTransactionByHex(ctx, c, txHex)
		if err != nil {
			return nil, err
		}

		if tx != nil {
			rts = &internalIncomingTx{
				Tx:           tx,
				BroadcastNow: false,
			}
		} else {
			rts = &externalIncomingTx{
				Hex:          txHex,
				BroadcastNow: false,
			}
		}
	}

	if err := rts.Validate(); err != nil {
		return nil, err
	}

	return rts, nil
}

func getTransactionByHex(ctx context.Context, c ClientInterface, hex string) (*Transaction, error) {
	// @arkadiusz: maybe we should actually search by hex?
	btTx, err := bt.NewTxFromString(hex)
	if err != nil {
		return nil, err
	}

	// Get the transaction by ID
	transaction, err := getTransactionByID(
		ctx, "", btTx.GetTxID(), c.DefaultModelOptions()...,
	)

	return transaction, err
}

func waitForRecordTxWriteLock(ctx context.Context, c ClientInterface, key string) func() {
	var (
		unlock func()
		err    error
	)
	// Create the lock and set the release for after the function completes
	// Waits for the moment when the transaction is unlocked and creates a new lock
	// Relevant for bux to bux transactions, as we have 1 tx but need to record 2 txs - outgoing and incoming
	for {
		unlock, err = newWriteLock(
			ctx, fmt.Sprintf(lockKeyRecordTx, key), c.Cachestore(),
		)
		if err == nil {
			break
		}
		time.Sleep(time.Second * 1)
	}

	return unlock
}
