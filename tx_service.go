package bux

import (
	"context"
	"errors"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-datastore"
)

// processTransactions will process transaction records
func processTransactions(ctx context.Context, maxTransactions int, opts ...ModelOps) error {
	queryParams := &datastore.QueryParams{
		Page:          1,
		PageSize:      maxTransactions,
		OrderByField:  "created_at",
		SortDirection: "asc",
	}

	conditions := map[string]interface{}{
		"$or": []map[string]interface{}{{
			blockHeightField: 0,
		}, {
			blockHeightField: nil,
		}},
	}

	records := make([]Transaction, 0)
	err := getModelsByConditions(ctx, ModelTransaction, &records, nil, &conditions, queryParams, opts...)
	if err != nil {
		return err
	} else if len(records) == 0 {
		return nil
	}

	txs := make([]*Transaction, 0)
	for index := range records {
		records[index].enrich(ModelTransaction, opts...)
		txs = append(txs, &records[index])
	}

	for index := range records {
		if err = _processTransaction(
			ctx, txs[index],
		); err != nil {
			return err
		}
	}

	return nil
}

// processUtxos will process the inputs and outputs for UTXOs
func (m *Transaction) processUtxos(ctx context.Context) error {
	// Input should be processed only for outgoing transactions
	if m.draftTransaction != nil {
		if err := m.processInputs(ctx); err != nil {
			return err
		}
	}

	return m._processOutputs(ctx)
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

// _processTransaction will process the sync transaction record, or save the failure
func _processTransaction(ctx context.Context, transaction *Transaction) error {
	txInfo, err := transaction.Client().Chainstate().QueryTransactionFastest(
		ctx, transaction.ID, chainstate.RequiredOnChain, defaultQueryTxTimeout,
	)
	if err != nil {
		if errors.Is(err, chainstate.ErrTransactionNotFound) {
			return nil
		}
		return err
	}

	transaction.BlockHash = txInfo.BlockHash
	transaction.BlockHeight = uint64(txInfo.BlockHeight)

	return transaction.Save(ctx)
}