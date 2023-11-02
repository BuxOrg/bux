package bux

import (
	"context"
	"errors"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

type outgoingTx struct {
	Hex            string
	RelatedDraftID string
	XPubKey        string
}

func (tx *outgoingTx) Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	// process
	transaction, err := _createOutgoingTxToRecord(ctx, tx, c, opts)

	if err != nil {
		return nil, fmt.Errorf("OutgoingTx.Execute(): creation of outgoing tx failed. Reason: %w", err)
	}

	if transaction.syncTransaction.P2PStatus == SyncStatusReady {
		if err = _processP2PTransaction(ctx, transaction.syncTransaction, transaction); err != nil {
			transaction.client.Logger().
				Error(ctx, fmt.Sprintf("OutgoingTx.Execute(): processP2PTransaction failed. Reason: %s", err)) // TODO: add transaction info to log context

			return nil, err // reject transaction if P2P notification failed
		}
	}

	if transaction.syncTransaction.BroadcastStatus == SyncStatusReady {
		if err = broadcastSyncTransaction(ctx, transaction.syncTransaction); err != nil {
			// ignore error, transaction will be broadcasted by cron task
			transaction.client.Logger().
				Warn(ctx, fmt.Sprintf("OutgoingTx.Execute(): broadcasting failed. Reason: %s", err)) // TODO: add transaction info to log context
		}
	}

	// record
	if err = transaction.Save(ctx); err != nil {
		return nil, fmt.Errorf("OutgoingTx.Execute(): saving of Transaction failed. Reason: %w", err)
	}

	return transaction, nil
}

func (tx outgoingTx) Validate() error {
	if tx.Hex == "" {
		return ErrMissingFieldHex
	}

	if tx.RelatedDraftID == "" {
		return errors.New("empty RelatedDraftID")
	}

	if tx.XPubKey == "" {
		return errors.New("empty xPubKey") // is it required ?
	}

	return nil // is valid
}

func (tx outgoingTx) TxID() string {
	btTx, _ := bt.NewTxFromString(tx.Hex)
	return btTx.TxID()
}

func _createOutgoingTxToRecord(ctx context.Context, oTx *outgoingTx, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	// Create NEW transaction model
	newOpts := c.DefaultModelOptions(append(opts, WithXPub(oTx.XPubKey), New())...)
	tx := newTransactionWithDraftID(
		oTx.Hex, oTx.RelatedDraftID, newOpts...,
	)

	// hydrate
	if err := _hydrateOutgoingWithDraft(ctx, tx); err != nil {
		return nil, err
	}

	_hydrateOutgoingWithSync(tx)

	if err := tx.processUtxos(ctx); err != nil {
		return nil, err
	}

	tx.TotalValue, tx.Fee = tx.getValues()
	if tx.TransactionBase.parsedTx != nil {
		tx.NumberOfInputs = uint32(len(tx.TransactionBase.parsedTx.Inputs))
		tx.NumberOfOutputs = uint32(len(tx.TransactionBase.parsedTx.Outputs))
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
	//sync.SyncStatus = SyncStatusReady

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
				broadcast = SyncStatusPending // postpone broadcasting if tx contains outputs in BEEF

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
			p2pStatus = SyncStatusPending

			break
		}
	}

	return p2pStatus
}
