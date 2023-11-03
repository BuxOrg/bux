package bux

import (
	"context"
	"errors"
	"fmt"
)

type internalIncomingTx struct {
	Tx           *Transaction
	BroadcastNow bool // e.g. BEEF must be broadcasted now
}

func (tx *internalIncomingTx) Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	transaction := tx.Tx

	// process
	syncTx, err := GetSyncTransactionByID(ctx, transaction.ID, transaction.GetOptions(false)...)
	if err != nil {
		return nil, fmt.Errorf("InternalIncomingTx.Execute(): getting syncTx failed. Reason: %w", err)
	}

	if tx.BroadcastNow || syncTx.BroadcastStatus == SyncStatusReady {
		syncTx.transaction = transaction
		err := broadcastSyncTransaction(ctx, syncTx)
		if err != nil {
			transaction.client.Logger().
				Warn(ctx, fmt.Sprintf("InternalIncomingTx.Execute(): broadcasting failed. Reason: %s", err))

			if syncTx.BroadcastStatus == SyncStatusSkipped { // revert status to ready after fail to re-run broadcasting, this can happen when we received internal BEEF tx
				syncTx.BroadcastStatus = SyncStatusReady

				if err = syncTx.Save(ctx); err != nil {
					transaction.client.Logger().
						Error(ctx, fmt.Sprintf("InternalIncomingTx.Execute(): changing synctx.BroadcastStatus from Pending to Ready failed. Reason: %s", err))
				}
			}

			// ignore broadcast error - will be repeted by task manager
		}
	}

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
