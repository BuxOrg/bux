package bux

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
)

type internalIncomingTx struct {
	Tx                   *Transaction
	broadcastNow         bool // e.g. BEEF must be broadcasted now
	allowBroadcastErrors bool // only BEEF cannot allow for broadcast errors
}

func (strategy *internalIncomingTx) Execute(ctx context.Context, c ClientInterface, _ []ModelOps) (*Transaction, error) {
	logger := c.Logger()
	logger.Info().
		Str("txID", strategy.Tx.ID).
		Msgf("InternalIncomingTx.Execute(): start, TxID: %s", strategy.Tx.ID)

	// process
	transaction := strategy.Tx
	syncTx, err := GetSyncTransactionByID(ctx, transaction.ID, transaction.GetOptions(false)...)
	if err != nil || syncTx == nil {
		return nil, fmt.Errorf("InternalIncomingTx.Execute(): getting syncTx failed. Reason: %w", err)
	}

	if strategy.broadcastNow || syncTx.BroadcastStatus == SyncStatusReady {
		syncTx.transaction = transaction
		transaction.syncTransaction = syncTx

		err := _internalIncomingBroadcast(ctx, logger, transaction, strategy.allowBroadcastErrors)
		if err != nil {
			logger.Error().
				Str("txID", transaction.ID).
				Msgf("InternalIncomingTx.Execute(): broadcasting failed, transaction rejected! Reason: %s, TxID: %s", err, transaction.ID)

			return nil, fmt.Errorf("InternalIncomingTx.Execute(): broadcasting failed, transaction rejected! Reason: %w, TxID: %s", err, transaction.ID)
		}
	}

	logger.Info().
		Str("txID", transaction.ID).
		Msgf("InternalIncomingTx.Execute(): complete, TxID: %s", transaction.ID)
	return transaction, nil
}

func (strategy *internalIncomingTx) Validate() error {
	if strategy.Tx == nil {
		return errors.New("Tx cannot be nil")
	}

	return nil // is valid
}

func (strategy *internalIncomingTx) TxID() string {
	return strategy.Tx.ID
}

func (strategy *internalIncomingTx) LockKey() string {
	return fmt.Sprintf("incoming-%s", strategy.Tx.ID)
}

func (strategy *internalIncomingTx) ForceBroadcast(force bool) {
	strategy.broadcastNow = force
}

func (strategy *internalIncomingTx) FailOnBroadcastError(forceFail bool) {
	strategy.allowBroadcastErrors = !forceFail
}

func _internalIncomingBroadcast(ctx context.Context, logger *zerolog.Logger, transaction *Transaction, allowErrors bool) error {
	logger.Info().
		Str("txID", transaction.ID).
		Msgf("InternalIncomingTx.Execute(): start broadcast, TxID: %s", transaction.ID)

	syncTx := transaction.syncTransaction
	err := broadcastSyncTransaction(ctx, syncTx)

	if err == nil {
		logger.Info().
			Str("txID", transaction.ID).
			Msgf("InternalIncomingTx.Execute(): broadcast complete, TxID: %s", transaction.ID)

		return nil
	}

	if allowErrors {
		logger.Warn().
			Str("txID", transaction.ID).
			Msgf("InternalIncomingTx.Execute(): broadcasting failed, next try will be handled by task manager. Reason: %s, TxID: %s", err, transaction.ID)

		// ignore broadcast error - will be repeted by task manager
		return nil
	}

	return err
}
