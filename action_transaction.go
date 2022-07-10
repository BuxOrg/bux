package bux

import (
	"context"
	"fmt"

	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-datastore"
)

// RecordTransaction will parse the transaction and save it into the Datastore
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
		ctx, fmt.Sprintf(lockKeyRecordTx, id), c.Cachestore(),
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

			// Create the sync transaction model
			sync := newSyncTransaction(
				transaction.GetID(),
				transaction.Client().DefaultSyncConfig(),
				transaction.GetOptions(true)...,
			)

			// Skip broadcasting and skip P2P (incoming tx should have been broadcasted already)
			sync.BroadcastStatus = SyncStatusSkipped // todo: this is an assumption
			sync.P2PStatus = SyncStatusSkipped       // The owner of the Tx should have already notified paymail providers

			// Use the same metadata
			sync.Metadata = transaction.Metadata

			// If all the options are skipped, do not make a new model (ignore the record)
			if !sync.isSkipped() {
				if err = sync.Save(ctx); err != nil {
					return nil, err
				}
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

// RecordRawTransaction will parse the transaction and save it into the Datastore directly, without any checks
//
// Only use this function when you know what you are doing!
//
// txHex is the raw transaction hex
// opts are model options and can include "metadata"
func (c *Client) RecordRawTransaction(ctx context.Context, txHex string,
	opts ...ModelOps) (*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "record_raw_transaction")

	return c.recordTxHex(ctx, txHex, opts...)
}

// RecordMonitoredTransaction will parse the transaction and save it into the Datastore
//
// This function will try to record the transaction directly, without checking draft ids etc.
//nolint:nolintlint,unparam // opts is the way, but not yet being used
func recordMonitoredTransaction(ctx context.Context, client ClientInterface, txHex string,
	opts ...ModelOps) (*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = client.GetOrStartTxn(ctx, "record_monitored_transaction")

	return client.recordTxHex(ctx, txHex, opts...)
}

func (c *Client) recordTxHex(ctx context.Context, txHex string, opts ...ModelOps) (*Transaction, error) {
	// Create the model & set the default options (gives options from client->model)
	newOpts := c.DefaultModelOptions(append(opts, New())...)
	transaction := newTransaction(txHex, newOpts...)

	// Ensure that we have a transaction id (created from the txHex)
	id := transaction.GetID()
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
		ctx, fmt.Sprintf(lockKeyProcessXpub, utils.Hash(rawXpubKey)), c.Cachestore(),
	)
	defer unlock()
	if err != nil {
		return nil, err
	}

	// Create the draft tx model
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
func (c *Client) GetTransaction(ctx context.Context, xPubID, txID string) (*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_transaction")

	// Get the transaction by ID
	transaction, err := getTransactionByID(
		ctx, xPubID, txID, c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, ErrMissingTransaction
	}

	return transaction, nil
}

// GetTransactions will get all the transactions from the Datastore
func (c *Client) GetTransactions(ctx context.Context, metadataConditions *Metadata,
	conditions *map[string]interface{}, queryParams *datastore.QueryParams, opts ...ModelOps) ([]*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_transactions")

	// Get the transactions
	transactions, err := getTransactions(
		ctx, metadataConditions, conditions, queryParams,
		c.DefaultModelOptions(opts...)...,
	)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetTransactionsAggregate will get a count of all transactions per aggregate column from the Datastore
func (c *Client) GetTransactionsAggregate(ctx context.Context, metadataConditions *Metadata,
	conditions *map[string]interface{}, aggregateColumn string, opts ...ModelOps) (map[string]interface{}, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_transactions")

	// Get the transactionAggregate
	transactionAggregate, err := getTransactionsAggregate(
		ctx, metadataConditions, conditions, aggregateColumn,
		c.DefaultModelOptions(opts...)...,
	)
	if err != nil {
		return nil, err
	}

	return transactionAggregate, nil
}

// GetTransactionsCount will get a count of all the transactions from the Datastore
func (c *Client) GetTransactionsCount(ctx context.Context, metadataConditions *Metadata,
	conditions *map[string]interface{}, opts ...ModelOps) (int64, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_transactions_count")

	// Get the transactions count
	count, err := getTransactionsCount(
		ctx, metadataConditions, conditions,
		c.DefaultModelOptions(opts...)...,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetTransactionsByXpubID will get all transactions for a given xpub from the Datastore
//
// ctx is the context
// rawXpubKey is the raw xPub key
// metadataConditions is added to the request for searching
// conditions is added the request for searching
func (c *Client) GetTransactionsByXpubID(ctx context.Context, xPubID string, metadataConditions *Metadata,
	conditions *map[string]interface{}, queryParams *datastore.QueryParams) ([]*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "get_transaction")

	// Get the transaction by ID
	// todo: add queryParams for: page size and page (right now it is unlimited)
	transactions, err := getTransactionsByXpubID(
		ctx, xPubID, metadataConditions, conditions, queryParams,
		c.DefaultModelOptions()...,
	)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetTransactionsByXpubIDCount will get the count of all transactions matching the search criteria
func (c *Client) GetTransactionsByXpubIDCount(ctx context.Context, xPubID string, metadataConditions *Metadata,
	conditions *map[string]interface{}) (int64, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "count_transactions")

	count, err := getTransactionsCountByXpubID(
		ctx, xPubID, metadataConditions, conditions,
		c.DefaultModelOptions()...,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateTransactionMetadata will update the metadata in an existing transaction
func (c *Client) UpdateTransactionMetadata(ctx context.Context, xPubID, id string,
	metadata Metadata) (*Transaction, error) {

	// Check for existing NewRelic transaction
	ctx = c.GetOrStartTxn(ctx, "update_transaction_by_id")

	// Get the transaction
	transaction, err := c.GetTransaction(ctx, xPubID, id)
	if err != nil {
		return nil, err
	}

	// Update the metadata
	if err = transaction.UpdateTransactionMetadata(
		xPubID, metadata,
	); err != nil {
		return nil, err
	}

	// Save the model
	if err = transaction.Save(ctx); err != nil {
		return nil, err
	}

	return transaction, nil
}
