package bux

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/BuxOrg/bux/utils"
)

func registerRawTransaction(ctx context.Context, c ClientInterface, allowUnknown bool, txHex string, opts ...ModelOps) (*Transaction, error) {
	newOpts := c.DefaultModelOptions(append(opts, New())...)
	tx := newTransaction(txHex, newOpts...)

	// Ensure that we have a transaction id (created from the txHex)
	id := tx.GetID()
	if len(id) == 0 {
		return nil, ErrMissingTxHex
	}

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, fmt.Sprintf(lockKeyRecordTx, id), c.Cachestore(),
	)
	defer unlock()
	if err != nil {
		return nil, err
	}

	// Logic moved from BeforeCreating hook - should be refactorized in next iteration

	if !c.IsITCEnabled() && !tx.hasOneKnownDestination(ctx, c) {
		return nil, ErrNoMatchingOutputs
	}

	// Process the UTXOs
	if err = tx.processUtxos(ctx); err != nil {
		return nil, err
	}

	// Set the values from the inputs/outputs and draft tx
	tx.TotalValue, tx.Fee = tx.getValues()

	// Add values
	tx.NumberOfInputs = uint32(len(tx.parsedTx.Inputs))
	tx.NumberOfOutputs = uint32(len(tx.parsedTx.Outputs))

	// /Logic moved from BeforeCreating hook - should be refactorized in next iteration

	// do not register transactions we have nothing to do with (this check must be done after transaction.processUtxos())
	if !allowUnknown && tx.XpubInIDs == nil && tx.XpubOutIDs == nil {
		return nil, ErrTransactionUnknown
	}

	if !tx.isMined() {
		sync := newSyncTransaction(
			tx.GetID(),
			c.DefaultSyncConfig(),
			tx.GetOptions(true)...,
		)
		sync.BroadcastStatus = SyncStatusSkipped
		sync.P2PStatus = SyncStatusSkipped

		sync.Metadata = tx.Metadata
		tx.syncTransaction = sync
	}

	if err = tx.Save(ctx); err != nil {
		return nil, err
	}

	return tx, nil
}

// processUtxos will process the inputs and outputs for UTXOs
func (m *Transaction) processUtxos(ctx context.Context) error {
	// Input should be processed only for outgoing transactions
	if m.draftTransaction != nil {
		if err := m._processInputs(ctx); err != nil {
			return err
		}
	}

	return m._processOutputs(ctx)
}

// processTxInputs will process the transaction inputs
func (m *Transaction) _processInputs(ctx context.Context) (err error) {
	// Pre-build the options
	opts := m.GetOptions(false)
	client := m.Client()

	var utxo *Utxo

	// check whether we are spending an internal utxo
	for index := range m.TransactionBase.parsedTx.Inputs {
		// todo: optimize this SQL SELECT to get all utxos in one query?
		if utxo, err = m.transactionService.getUtxo(ctx,
			hex.EncodeToString(m.TransactionBase.parsedTx.Inputs[index].PreviousTxID()),
			m.TransactionBase.parsedTx.Inputs[index].PreviousTxOutIndex,
			opts...,
		); err != nil {
			return
		} else if utxo != nil { // Found a UTXO record

			// Is Spent?
			if len(utxo.SpendingTxID.String) > 0 {
				return ErrUtxoAlreadySpent
			}

			// Only if IUC is enabled (or client is nil which means its enabled by default)
			if client == nil || client.IsIUCEnabled() {

				// check whether the utxo is spent
				isReserved := len(utxo.DraftID.String) > 0
				matchesDraft := m.draftTransaction != nil && utxo.DraftID.String == m.draftTransaction.ID

				// Check whether the spending transaction was reserved by the draft transaction (in the utxo)
				if !isReserved {
					return ErrUtxoNotReserved
				}
				if !matchesDraft {
					return ErrDraftIDMismatch
				}
			}

			// Update the output value
			if _, ok := m.XpubOutputValue[utxo.XpubID]; !ok {
				m.XpubOutputValue[utxo.XpubID] = 0
			}
			m.XpubOutputValue[utxo.XpubID] -= int64(utxo.Satoshis)

			// Mark utxo as spent
			utxo.SpendingTxID.Valid = true
			utxo.SpendingTxID.String = m.ID
			m.utxos = append(m.utxos, *utxo)

			// Add the xPub ID
			if !utils.StringInSlice(utxo.XpubID, m.XpubInIDs) {
				m.XpubInIDs = append(m.XpubInIDs, utxo.XpubID)
			}
		}

		// todo: what if the utxo is nil (not found)?
	}

	return
}

// processTxOutputs will process the transaction outputs
func (m *Transaction) _processOutputs(ctx context.Context) (err error) {
	// Pre-build the options
	opts := m.GetOptions(false)
	newOpts := append(opts, New())
	var destination *Destination

	// check all the outputs for a known destination
	numberOfOutputsProcessed := 0
	for index := range m.TransactionBase.parsedTx.Outputs {
		amount := m.TransactionBase.parsedTx.Outputs[index].Satoshis

		// only save outputs with a satoshi value attached to it
		if amount > 0 {

			txLockingScript := m.TransactionBase.parsedTx.Outputs[index].LockingScript.String()
			lockingScript := utils.GetDestinationLockingScript(txLockingScript)

			// only Save utxos for known destinations
			// todo: optimize this SQL SELECT by requesting all the scripts at once (vs in this loop)
			// todo: how to handle tokens and other non-standard outputs ?
			if destination, err = m.transactionService.getDestinationByLockingScript(
				ctx, lockingScript, opts...,
			); err != nil {
				return
			} else if destination != nil {

				// Add value of output to xPub ID
				if _, ok := m.XpubOutputValue[destination.XpubID]; !ok {
					m.XpubOutputValue[destination.XpubID] = 0
				}
				m.XpubOutputValue[destination.XpubID] += int64(amount)

				utxo, _ := m.client.GetUtxoByTransactionID(ctx, m.ID, uint32(index))
				if utxo == nil {
					utxo = newUtxo(
						destination.XpubID, m.ID, txLockingScript, uint32(index),
						amount, newOpts...,
					)
				}
				// Append the UTXO model
				m.utxos = append(m.utxos, *utxo)

				// Add the xPub ID
				if !utils.StringInSlice(destination.XpubID, m.XpubOutIDs) {
					m.XpubOutIDs = append(m.XpubOutIDs, destination.XpubID)
				}

				numberOfOutputsProcessed++
			}
		}
	}

	return
}
