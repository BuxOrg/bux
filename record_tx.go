package bux

import (
	"context"
	"fmt"
	"time"
)

type recordTxStrategy interface {
	TxID() string
	LockKey() string
	Validate() error
	Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error)
}

type recordIncomingTxStrategy interface {
	recordTxStrategy
	ForceBroadcast(force bool)
	FailOnBroadcastError(forceFail bool)
}

func recordTransaction(ctx context.Context, c ClientInterface, strategy recordTxStrategy, opts ...ModelOps) (*Transaction, error) {
	unlock := waitForRecordTxWriteLock(ctx, c, strategy.LockKey())
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

func getIncomingTxRecordStrategy(ctx context.Context, c ClientInterface, txHex string) (recordIncomingTxStrategy, error) {
	tx, err := getTransactionByHex(ctx, txHex, c.DefaultModelOptions()...)
	if err != nil {
		return nil, err
	}

	var rts recordIncomingTxStrategy

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

	lockKey := fmt.Sprintf(lockKeyRecordTx, key)

	// TODO: change to DEBUG level log when we will support it
	c.Logger().Info(ctx, fmt.Sprintf("try add write lock %s", lockKey))

	for {

		unlock, err = newWriteLock(
			ctx, lockKey, c.Cachestore(),
		)
		if err == nil {
			// TODO: change to DEBUG level log when we will support it
			c.Logger().Info(ctx, fmt.Sprintf("added write lock %s", lockKey))
			break
		}
		time.Sleep(time.Second * 1)
	}

	return func() {
		// TODO: change to DEBUG level log when we will support it
		c.Logger().Info(ctx, fmt.Sprintf("unlock %s", lockKey))
		unlock()
	}
}
