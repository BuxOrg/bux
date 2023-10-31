package bux

import (
	"context"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

type externalIncomingTx struct {
	Hex          string
	BroadcastNow bool // e.g. BEEF must be broadcasted now
}

func (tx *externalIncomingTx) Execute(ctx context.Context, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	// process
	if !tx.BroadcastNow && c.IsITCEnabled() { // do not save transaction to database now, save IncomingTransaction instead and let task manager handle and process it
		return _addTxToCheck(ctx, tx, c, opts)
	}

	transaction, err := _createExternalTxToRecord(ctx, tx, c, opts)

	if err != nil {
		return nil, fmt.Errorf("ExternalIncomingTx.Execute(): creation of external incoming tx failed. Reason: %w", err)
	}

	if transaction.syncTransaction.BroadcastStatus == SyncStatusReady {
		if err = broadcastSyncTransaction(ctx, transaction.syncTransaction); err != nil {
			// ignore error, transaction will be broadcaset by cron task - @arkadiusz: to chyba nie do końca prawda
			transaction.client.Logger().
				Warn(ctx, fmt.Sprintf("ExternalIncomingTx.Execute(): broadcasting failed. Reason: %s", err)) // TODO: add transaction info to log context
		}
	}

	// record
	if err = transaction.Save(ctx); err != nil {
		return nil, fmt.Errorf("ExternalIncomingTx.Execute(): saving of Transaction failed. Reason: %w", err)
	}

	return transaction, nil
}

func (tx *externalIncomingTx) Validate() error {
	if tx.Hex == "" {
		return ErrMissingFieldHex
	}

	return nil // is valid
}

func (tx *externalIncomingTx) TxID() string {
	btTx, _ := bt.NewTxFromString(tx.Hex)
	return btTx.TxID()
}

func (tx *externalIncomingTx) ForceBroadcast(force bool) {
	tx.BroadcastNow = force
}

func _addTxToCheck(ctx context.Context, tx *externalIncomingTx, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	incomingTx := newIncomingTransaction(tx.Hex, c.DefaultModelOptions(append(opts, New())...)...)

	if err := incomingTx.Save(ctx); err != nil {
		return nil, fmt.Errorf("ExternalIncomingTx.Execute(): addind new IncomingTx to check queue failed. Reason: %w", err)
	}

	// TODO: ensure I don't need syncTransaction here

	return newTransactionFromIncomingTransaction(incomingTx), nil // TODO: change incoming processing
}

func _createExternalTxToRecord(ctx context.Context, eTx *externalIncomingTx, c ClientInterface, opts []ModelOps) (*Transaction, error) {
	// Create NEW tx model
	tx := newTransaction(eTx.Hex, c.DefaultModelOptions(append(opts, New())...)...)
	_hydrateExternalWithSync(tx)

	// Check that the transaction has >= 1 known destination
	if !tx.TransactionBase.hasOneKnownDestination(ctx, c, tx.GetOptions(false)...) {
		return nil, ErrNoMatchingOutputs
	}

	// Process the UTXOs
	if err := tx.processUtxos(ctx); err != nil {
		return nil, err
	}

	// Set the values from the inputs/outputs and draft tx
	tx.TotalValue, tx.Fee = tx.getValues()

	// Add values if found
	if tx.TransactionBase.parsedTx != nil {
		tx.NumberOfInputs = uint32(len(tx.TransactionBase.parsedTx.Inputs))
		tx.NumberOfOutputs = uint32(len(tx.TransactionBase.parsedTx.Outputs))
	}

	return tx, nil
}

func _hydrateExternalWithSync(tx *Transaction) {
	// Create the sync transaction model
	sync := newSyncTransaction(
		tx.ID,
		tx.Client().DefaultSyncConfig(),
		tx.GetOptions(true)...,
	)

	// to simplfy: broadcast every external incoming txs
	sync.BroadcastStatus = SyncStatusReady

	sync.P2PStatus = SyncStatusSkipped // The owner of the Tx should have already notified paymail providers
	//sync.SyncStatus = SyncStatusReady

	// Use the same metadata
	sync.Metadata = tx.Metadata

	// @arkadiusz: my assumptium is we cannot skip sync here
	// // If all the options are skipped, do not make a new model (ignore the record)
	// if !sync.isSkipped() {
	// 	m.syncTransaction = sync
	// }

	sync.transaction = tx
	tx.syncTransaction = sync
}
