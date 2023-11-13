package bux

import (
	"context"
	"errors"
	"fmt"

	zLogger "github.com/mrz1836/go-logger"
)

type internalIncomingTx struct {
	Tx           *Transaction
	BroadcastNow bool // e.g. BEEF must be broadcasted now
}

func (tx *internalIncomingTx) Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	logger := c.Logger()
	logger.Info(ctx, fmt.Sprintf("InternalIncomingTx.Execute(): start, TxID: %s", tx.Tx.ID))

	// process
	transaction := tx.Tx
	syncTx, err := GetSyncTransactionByID(ctx, transaction.ID, transaction.GetOptions(false)...)
	if err != nil {
		return nil, fmt.Errorf("InternalIncomingTx.Execute(): getting syncTx failed. Reason: %w", err)
	}

	if tx.BroadcastNow || syncTx.BroadcastStatus == SyncStatusReady {
		syncTx.transaction = transaction
		transaction.syncTransaction = syncTx

		_internalIncomingBroadcast(ctx, logger, transaction) // ignore broadcast error - will be repeted by task manager
	}

	logger.Info(ctx, fmt.Sprintf("InternalIncomingTx.Execute(): complete, TxID: %s", transaction.ID))
	return transaction, nil
}

func (tx *internalIncomingTx) Validate() error {
	if tx.Tx == nil {
		return errors.New("Tx cannot be nil")
	}

	return nil // is valid
}

func (tx *internalIncomingTx) TxID() string {
	return tx.Tx.ID
}

func (tx *internalIncomingTx) ForceBroadcast(force bool) {
	tx.BroadcastNow = force
}

func _internalIncomingBroadcast(ctx context.Context, logger zLogger.GormLoggerInterface, transaction *Transaction) {
	logger.Info(ctx, fmt.Sprintf("InternalIncomingTx.Execute(): start broadcast, TxID: %s", transaction.ID))

	syncTx := transaction.syncTransaction
	err := broadcastSyncTransaction(ctx, syncTx)
	if err != nil {
		logger.
			Warn(ctx, fmt.Sprintf("InternalIncomingTx.Execute(): broadcasting failed. Reason: %s, TxID: %s", err, transaction.ID))

		if syncTx.BroadcastStatus == SyncStatusSkipped { // revert status to ready after fail to re-run broadcasting, this can happen when we received internal BEEF tx
			syncTx.BroadcastStatus = SyncStatusReady

			if err = syncTx.Save(ctx); err != nil {
				logger.
					Error(ctx, fmt.Sprintf("InternalIncomingTx.Execute(): changing synctx.BroadcastStatus from Pending to Ready failed. Reason: %s, TxID: %s", err, transaction.ID))
			}
		}

		// ignore broadcast error - will be repeted by task manager
	} else {
		logger.
			Info(ctx, fmt.Sprintf("InternalIncomingTx.Execute(): broadcast complete, TxID: %s", transaction.ID))
	}
}
