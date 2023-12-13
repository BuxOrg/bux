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
		Msg("start recording transaction")

	// process
	transaction := strategy.Tx
	syncTx, err := GetSyncTransactionByID(ctx, transaction.ID, transaction.GetOptions(false)...)
	if err != nil || syncTx == nil {
		return nil, fmt.Errorf("getting syncTx failed. Reason: %w", err)
	}

	if strategy.broadcastNow || syncTx.BroadcastStatus == SyncStatusReady {
		syncTx.transaction = transaction
		transaction.syncTransaction = syncTx

		err := _internalIncomingBroadcast(ctx, logger, transaction, strategy.allowBroadcastErrors)
		if err != nil {
			logger.Error().
				Str("txID", transaction.ID).
				Msgf("broadcasting failed, transaction rejected! Reason: %s", err)

			return nil, fmt.Errorf("broadcasting failed, transaction rejected! Reason: %w", err)
		}
	}

	logger.Info().
		Str("txID", transaction.ID).
		Msg("complete")
	return transaction, nil
}

func (strategy *internalIncomingTx) Validate() error {
	if strategy.Tx == nil {
		return errors.New("tx cannot be nil")
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
		Msg("start broadcast")

	syncTx := transaction.syncTransaction
	err := broadcastSyncTransaction(ctx, syncTx)

	if err == nil {
		logger.Info().
			Str("txID", transaction.ID).
			Msg("broadcast complete")

		return nil
	}

	if allowErrors {
		logger.Warn().
			Str("txID", transaction.ID).
			Msgf("broadcasting failed, next try will be handled by task manager. Reason: %s", err)

		// ignore broadcast error - will be repeted by task manager
		return nil
	}

	return err
}
