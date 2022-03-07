package bux

import (
	"context"

	"github.com/BuxOrg/bux/utils"
)

// RecordTransaction will parse the transaction and Save it into the Datastore
//
// Internal (known) transactions: there is a corresponding `draft_transaction` via `draft_id`
// External (known) transactions: there are output(s) related to the destination `reference_id`, tx is valid (mempool/on-chain)
// External (unknown) transactions: no reference id but some output(s) match known outputs, tx is valid (mempool/on-chain)
// Unknown transactions: no matching outputs, tx will be disregarded
//
// xPubKey is the raw public xPub
// txHex is the raw transaction hex
// draftID is the unique draft id from a previously started New() transaction (draft_transaction.ID)
// opts are model options and can include "metadata"
func (c *Client) RecordTransaction(ctx context.Context, xPubKey, txHex, draftID string,
	opts ...ModelOps) (*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "record_transaction")

	// Create the model & set the default options (gives options from client->model)
	newOpts := c.DefaultModelOptions(append(opts, WithXPub(xPubKey), New())...)
	transaction := newTransactionWithDraftID(
		txHex, draftID, newOpts...,
	)

	// Ensure that we have a transaction id (created from the txHex)
	id := transaction.GetID()
	if len(id) == 0 {
		return nil, ErrMissingTxHex
	}

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, "action-record-transaction-"+id, c.Cachestore(),
	)
	defer unlock()
	if err != nil {
		return nil, err
	}

	// OPTION: check incoming transactions (if enabled, will add to queue for checking on-chain)
	if !c.IsITCEnabled() {
		transaction.DebugLog("incoming transaction check is disabled")
	} else {

		// Incoming (external/unknown) transaction (no draft id was given)
		if len(transaction.DraftID) == 0 {

			// Process & save the model
			incomingTx := newIncomingTransaction(
				transaction.ID, txHex, newOpts...,
			)
			if err = incomingTx.Save(ctx); err != nil {
				return nil, err
			}

			// Added to queue
			return newTransactionFromIncomingTransaction(incomingTx), nil
		}

		// Internal tx (must match draft tx)
		if transaction.draftTransaction, err = getDraftTransactionID(
			ctx, transaction.xPubID, transaction.DraftID,
			transaction.GetOptions(false)...,
		); err != nil {
			return nil, err
		} else if transaction.draftTransaction == nil {
			return nil, ErrDraftNotFound
		}
	}

	// Process & save the transaction model
	if err = transaction.Save(ctx); err != nil {
		return nil, err
	}

	// Return the response
	return transaction, nil
}

// NewTransaction will create a new draft transaction and return it
//
// ctx is the context
// rawXpubKey is the raw xPub key
// config is the TransactionConfig
// metadata is added to the model
// opts are additional model options to be applied
func (c *Client) NewTransaction(ctx context.Context, rawXpubKey string, config *TransactionConfig,
	opts ...ModelOps) (*DraftTransaction, error) {

	// Check for existing NewRelic draftTransaction
	ctx = c.GetOrStartTxn(ctx, "new_transaction")

	// Create the lock and set the release for after the function completes
	unlock, err := newWaitWriteLock(
		ctx, "action-xpub-"+utils.Hash(rawXpubKey), c.Cachestore(),
	)
	defer unlock()
	if err != nil {
		return nil, err
	}

	// todo: this needs adjusting via Chainstate or mAPI
	if config.FeeUnit == nil {
		config.FeeUnit = c.GetFeeUnit(ctx, "miner")
	}

	// Create the model & set the default options (gives options from client->model)
	draftTransaction := newDraftTransaction(
		rawXpubKey, config,
		c.DefaultModelOptions(append(opts, New())...)...,
	)

	// Save the model
	if err = draftTransaction.Save(ctx); err != nil {
		return nil, err
	}

	// Return the created model
	return draftTransaction, nil
}

// GetTransaction will get a transaction from the Datastore
//
// ctx is the context
// testTxID is the transaction ID
func (c *Client) GetTransaction(ctx context.Context, rawXpubKey, txID string) (*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_transaction")

	// Get the transaction by ID
	transaction, err := getTransactionByID(
		ctx, rawXpubKey, txID, c.DefaultModelOptions(WithXPub(rawXpubKey))...,
	)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, ErrMissingTransaction
	}

	return transaction, nil
}

// GetTransactions will get all transactions for a given xpub from the Datastore
//
// ctx is the context
// rawXpubKey is the raw xPub key
// metadataConditions is added to the request for searching
// conditions is added the request for searching
func (c *Client) GetTransactions(ctx context.Context, rawXpubKey string, metadataConditions *Metadata,
	conditions *map[string]interface{}) ([]*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_transaction")

	// Get the transaction by ID
	// todo: add params for: page size and page (right now it is unlimited)
	transactions, err := getTransactionsByXpubID(
		ctx, rawXpubKey, metadataConditions, conditions, 0, 0,
		c.DefaultModelOptions(WithXPub(rawXpubKey))...,
	)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}
