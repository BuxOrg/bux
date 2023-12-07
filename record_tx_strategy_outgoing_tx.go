package bux

import (
	"context"
	"errors"
	"fmt"
	"github.com/libsv/go-bt/v2"
	"github.com/rs/zerolog"
)

type outgoingTx struct {
	Hex            string
	RelatedDraftID string
	XPubKey        string
}

func (strategy *outgoingTx) Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	logger := c.Logger()
	logger.Info().Msgf("OutgoingTx.Execute(): start, TxID: %s", strategy.TxID())

	// create
	transaction, err := _createOutgoingTxToRecord(ctx, strategy, c, opts)
	if err != nil {
		return nil, fmt.Errorf("OutgoingTx.Execute(): creation of outgoing tx failed. Reason: %w", err)
	}

	if err = transaction.Save(ctx); err != nil {
		return nil, fmt.Errorf("OutgoingTx.Execute(): saving of Transaction failed. Reason: %w", err)
	}

	// process
	if transaction.syncTransaction.P2PStatus == SyncStatusReady {
		if err = _outgoingNotifyP2p(ctx, logger, transaction); err != nil {
			// reject transaction if P2P notification failed
			logger.Error().Msgf("OutgoingTx.Execute(): transaction rejected by P2P provider, try to revert transaction. Reason: %s", err)

			if revertErr := c.RevertTransaction(ctx, transaction.ID); revertErr != nil {
				logger.Error().Msgf("OutgoingTx.Execute(): FATAL! Reverting transaction after failed P2P notification failed. Reason: %s", err)
			}

			return nil, err
		}
	}

	// get newest syncTx from DB - if it's an internal tx it could be broadcasted by us already
	syncTx, err := GetSyncTransactionByID(ctx, transaction.ID, transaction.GetOptions(false)...)
	if err != nil || syncTx == nil {
		return nil, fmt.Errorf("OutgoingTx.Execute(): getting syncTx failed. Reason: %w", err)
	}

	if syncTx.BroadcastStatus == SyncStatusReady {
		_outgoingBroadcast(ctx, logger, transaction) // ignore error
	}

	logger.Info().Msgf("OutgoingTx.Execute(): complete, TxID: %s", transaction.ID)
	return transaction, nil
}

func (strategy *outgoingTx) Validate() error {
	if strategy.Hex == "" {
		return ErrMissingFieldHex
	}

	if strategy.RelatedDraftID == "" {
		return errors.New("empty RelatedDraftID")
	}

	if strategy.XPubKey == "" {
		return errors.New("empty xPubKey")
	}

	return nil // is valid
}

func (strategy *outgoingTx) TxID() string {
	btTx, _ := bt.NewTxFromString(strategy.Hex)
	return btTx.TxID()
}

func (strategy *outgoingTx) LockKey() string {
	return fmt.Sprintf("outgoing-%s", strategy.TxID())
}

func _createOutgoingTxToRecord(ctx context.Context, oTx *outgoingTx, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	// Create NEW transaction model
	newOpts := c.DefaultModelOptions(append(opts, WithXPub(oTx.XPubKey), New())...)
	tx, err := newTransactionWithDraftID(
		oTx.Hex, oTx.RelatedDraftID, newOpts...,
	)

	if err != nil {
		return nil, err
	}

	// hydrate
	if err = _hydrateOutgoingWithDraft(ctx, tx); err != nil {
		return nil, err
	}

	_hydrateOutgoingWithSync(tx)

	if err := tx.processUtxos(ctx); err != nil {
		return nil, err
	}

	return tx, nil
}

func _hydrateOutgoingWithDraft(ctx context.Context, tx *Transaction) error {
	draft, err := getDraftTransactionID(ctx, tx.XPubID, tx.DraftID, tx.GetOptions(false)...)

	if err != nil {
		return err
	}

	if draft == nil {
		return ErrDraftNotFound
	}

	if len(draft.Configuration.Outputs) == 0 {
		return errors.New("corresponding draft transaction has no outputs")
	}

	if draft.Configuration.Sync == nil {
		draft.Configuration.Sync = tx.Client().DefaultSyncConfig()
	}

	tx.draftTransaction = draft

	return nil // success
}

func _hydrateOutgoingWithSync(tx *Transaction) {
	sync := newSyncTransaction(tx.ID, tx.draftTransaction.Configuration.Sync, tx.GetOptions(true)...)

	// setup synchronization
	sync.BroadcastStatus = _getBroadcastSyncStatus(tx)
	sync.P2PStatus = _getP2pSyncStatus(tx)
	sync.SyncStatus = SyncStatusPending // wait until transaction is broadcasted or P2P provider is notified

	sync.Metadata = tx.Metadata

	sync.transaction = tx
	tx.syncTransaction = sync
}

func _getBroadcastSyncStatus(tx *Transaction) SyncStatus {
	// immediately broadcast if is not BEEF
	broadcast := SyncStatusReady // broadcast immediately

	outputs := tx.draftTransaction.Configuration.Outputs

	for _, o := range outputs {
		if o.PaymailP4 != nil {
			if o.PaymailP4.Format == BeefPaymailPayloadFormat {
				broadcast = SyncStatusSkipped // postpone broadcasting if tx contains outputs in BEEF

				break
			}
		}
	}

	return broadcast
}

func _getP2pSyncStatus(tx *Transaction) SyncStatus {
	p2pStatus := SyncStatusSkipped

	outputs := tx.draftTransaction.Configuration.Outputs
	for _, o := range outputs {
		if o.PaymailP4 != nil && o.PaymailP4.ResolutionType == ResolutionTypeP2P {
			p2pStatus = SyncStatusReady // notify p2p immediately

			break
		}
	}

	return p2pStatus
}

func _outgoingNotifyP2p(ctx context.Context, logger *zerolog.Logger, tx *Transaction) error {
	logger.Info().Msgf("OutgoingTx.Execute(): start p2p, TxID: %s", tx.ID)

	if err := processP2PTransaction(ctx, tx.syncTransaction, tx); err != nil {
		logger.
			Error().Msgf("OutgoingTx.Execute(): processP2PTransaction failed. Reason: %s, TxID: %s", err, tx.ID)

		return err
	}

	logger.Info().Msgf("OutgoingTx.Execute(): p2p complete, TxID: %s", tx.ID)
	return nil
}

func _outgoingBroadcast(ctx context.Context, logger *zerolog.Logger, tx *Transaction) {
	logger.Info().Msgf("OutgoingTx.Execute(): start broadcast, TxID: %s", tx.ID)

	if err := broadcastSyncTransaction(ctx, tx.syncTransaction); err != nil {
		// ignore error, transaction will be broadcasted by cron task
		logger.
			Warn().Msgf("OutgoingTx.Execute(): broadcasting failed, next try will be handled by task manager. Reason: %s, TxID: %s", err, tx.ID)
	} else {
		logger.
			Info().Msgf("OutgoingTx.Execute(): broadcast complete, TxID: %s", tx.ID)
	}
}
