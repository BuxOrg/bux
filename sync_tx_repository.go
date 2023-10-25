package bux

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/libsv/go-bt/v2"
	"github.com/mrz1836/go-datastore"
)

/*** exported funcs ***/

// GetSyncTransactionByID will get a sync transaction
func GetSyncTransactionByID(ctx context.Context, id string, opts ...ModelOps) (*SyncTransaction, error) {
	// Get the records by status
	txs, err := _getSyncTransactionsByConditions(ctx,
		map[string]interface{}{
			idField: id,
		},
		nil, opts...,
	)
	if err != nil {
		return nil, err
	}
	if len(txs) != 1 {
		return nil, nil
	}

	return txs[0], nil
}

/*** /exported funcs ***/

/*** public unexported funcs ***/

// getTransactionsToBroadcast will get the sync transactions to broadcast
func getTransactionsToBroadcast(ctx context.Context, queryParams *datastore.QueryParams,
	opts ...ModelOps,
) (map[string][]*SyncTransaction, error) {
	// Get the records by status
	txs, err := _getSyncTransactionsByConditions(
		ctx,
		map[string]interface{}{
			broadcastStatusField: SyncStatusReady.String(),
		},
		queryParams, opts...,
	)
	if err != nil {
		return nil, err
	}

	// group transactions by xpub and return including the tx itself
	txsByXpub := make(map[string][]*SyncTransaction)
	for _, tx := range txs {
		if tx.transaction, err = getTransactionByID(
			ctx, "", tx.ID, opts...,
		); err != nil {
			return nil, err
		}

		var parentsBroadcast bool
		parentsBroadcast, err = _areParentsBroadcast(ctx, tx, opts...)
		if err != nil {
			return nil, err
		}

		if !parentsBroadcast {
			// if all parents are not broadcast, then we cannot broadcast this tx
			continue
		}

		xPubID := "" // fallback if we have no input xpubs
		if len(tx.transaction.XpubInIDs) > 0 {
			// use the first xpub for the grouping
			// in most cases when we are broadcasting, there should be only 1 xpub in
			xPubID = tx.transaction.XpubInIDs[0]
		}

		if txsByXpub[xPubID] == nil {
			txsByXpub[xPubID] = make([]*SyncTransaction, 0)
		}
		txsByXpub[xPubID] = append(txsByXpub[xPubID], tx)
	}

	return txsByXpub, nil
}

// getTransactionsToSync will get the sync transactions to sync
func getTransactionsToSync(ctx context.Context, queryParams *datastore.QueryParams,
	opts ...ModelOps,
) ([]*SyncTransaction, error) {
	// Get the records by status
	txs, err := _getSyncTransactionsByConditions(
		ctx,
		map[string]interface{}{
			syncStatusField: SyncStatusReady.String(),
		},
		queryParams, opts...,
	)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

// getTransactionsToNotifyP2P will get the sync transactions to notify p2p paymail providers
func getTransactionsToNotifyP2P(ctx context.Context, queryParams *datastore.QueryParams,
	opts ...ModelOps,
) ([]*SyncTransaction, error) {
	// Get the records by status
	txs, err := _getSyncTransactionsByConditions(
		ctx,
		map[string]interface{}{
			p2pStatusField: SyncStatusReady.String(),
		},
		queryParams, opts...,
	)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

/*** /public unexported funcs ***/

// getTransactionsToSync will get the sync transactions to sync
func _getSyncTransactionsByConditions(ctx context.Context, conditions map[string]interface{},
	queryParams *datastore.QueryParams, opts ...ModelOps,
) ([]*SyncTransaction, error) {
	if queryParams == nil {
		queryParams = &datastore.QueryParams{
			OrderByField:  createdAtField,
			SortDirection: datastore.SortAsc,
		}
	} else if queryParams.OrderByField == "" || queryParams.SortDirection == "" {
		queryParams.OrderByField = createdAtField
		queryParams.SortDirection = datastore.SortAsc
	}

	// Get the records
	var models []SyncTransaction
	if err := getModels(
		ctx, NewBaseModel(ModelNameEmpty, opts...).Client().Datastore(),
		&models, conditions, queryParams, defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	// Loop and enrich
	txs := make([]*SyncTransaction, 0)
	for index := range models {
		models[index].enrich(ModelSyncTransaction, opts...)
		txs = append(txs, &models[index])
	}

	return txs, nil
}

func _areParentsBroadcast(ctx context.Context, syncTx *SyncTransaction, opts ...ModelOps) (bool, error) {
	tx, err := getTransactionByID(ctx, "", syncTx.ID, opts...)
	if err != nil {
		return false, err
	}

	if tx == nil {
		return false, ErrMissingTransaction
	}

	// get the sync transaction of all inputs
	var btTx *bt.Tx
	btTx, err = bt.NewTxFromString(tx.Hex)
	if err != nil {
		return false, err
	}

	// check that all inputs we handled have been broadcast, or are not handled by Bux
	parentsBroadcast := true
	for _, input := range btTx.Inputs {
		var parentTx *SyncTransaction
		previousTxID := hex.EncodeToString(bt.ReverseBytes(input.PreviousTxID()))
		parentTx, err = GetSyncTransactionByID(ctx, previousTxID, opts...)
		if err != nil {
			return false, err
		}
		// if we have a sync transaction, and it is not complete, then we cannot broadcast
		if parentTx != nil && parentTx.BroadcastStatus != SyncStatusComplete {
			parentsBroadcast = false
		}
	}

	return parentsBroadcast, nil
}
