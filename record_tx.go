package bux

import (
	"context"
	"fmt"
	"time"
)

type recordTxStrategy interface {
	TxID() string
	Validate() error
	Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error)
}

type recordIncomingTxStrategy interface {
	ForceBroadcast(force bool)
	FailOnBroadcastError(forceFail bool)
}

func recordTransaction(ctx context.Context, c ClientInterface, strategy recordTxStrategy, opts ...ModelOps) (*Transaction, error) {
	unlock := waitForRecordTxWriteLock(ctx, c, strategy.TxID())
	defer unlock()

	return strategy.Execute(ctx, c, opts)
}

func getRecordTxStrategy(ctx context.Context, c ClientInterface, xPubKey, txHex, draftID string) (recordTxStrategy, error) {
	var rts recordTxStrategy

	if draftID != "" {
		rts = getOutgoingTxRecordStrategy(xPubKey, txHex, draftID)
	} else {
		var err error
		rts, err = getIncomingTxRecordStrategy(ctx, c, txHex)

		if err != nil {
			return nil, err
		}
	}

	if err := rts.Validate(); err != nil {
		return nil, err
	}

	return rts, nil
}

func getOutgoingTxRecordStrategy(xPubKey, txHex, draftID string) recordTxStrategy {
	return &outgoingTx{
		Hex:            txHex,
		RelatedDraftID: draftID,
		XPubKey:        xPubKey,
	}
}

func getIncomingTxRecordStrategy(ctx context.Context, c ClientInterface, txHex string) (recordTxStrategy, error) {
	tx, err := getTransactionByHex(ctx, txHex, c.DefaultModelOptions()...)
	if err != nil {
		return nil, err
	}

	var rts recordTxStrategy

	if tx != nil {
		rts = &internalIncomingTx{
			Tx:           tx,
			broadcastNow: false,
		}
	} else {
		rts = &externalIncomingTx{
			Hex:          txHex,
			broadcastNow: false,
		}
	}

	return rts, nil
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
