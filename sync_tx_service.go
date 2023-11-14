package bux

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/notifications"
	"github.com/bitcoin-sv/go-paymail"
	"github.com/mrz1836/go-datastore"
	customTypes "github.com/mrz1836/go-datastore/custom_types"
)

// processSyncTransactions will process sync transaction records
func processSyncTransactions(ctx context.Context, maxTransactions int, opts ...ModelOps) error {
	queryParams := &datastore.QueryParams{
		Page:          1,
		PageSize:      maxTransactions,
		OrderByField:  "created_at",
		SortDirection: "desc",
	}

	// Get x records
	records, err := getTransactionsToSync(
		ctx, queryParams, opts...,
	)
	if err != nil {
		return err
	} else if len(records) == 0 {
		return nil
	}

	// Process the incoming transaction
	for index := range records {
		if err = _syncTxDataFromChain(
			ctx, records[index], nil,
		); err != nil {
			return err
		}
	}

	return nil
}

// processBroadcastTransactions will process sync transaction records
func processBroadcastTransactions(ctx context.Context, maxTransactions int, opts ...ModelOps) error {
	queryParams := &datastore.QueryParams{
		Page:          1,
		PageSize:      maxTransactions,
		OrderByField:  createdAtField,
		SortDirection: datastore.SortAsc,
	}

	// Get maxTransactions records, grouped by xpub
	snTxs, err := getTransactionsToBroadcast(ctx, queryParams, opts...)
	if err != nil {
		return err
	} else if len(snTxs) == 0 {
		return nil
	}

	// Process the transactions per xpub, in parallel
	txsByXpub := _groupByXpub(snTxs)

	// we limit the number of concurrent broadcasts to the number of cpus*2, since there is lots of IO wait
	limit := make(chan bool, runtime.NumCPU()*2)
	wg := new(sync.WaitGroup)

	for xPubID := range txsByXpub {
		limit <- true // limit the number of routines running at the same time
		wg.Add(1)
		go func(xPubID string) {
			defer wg.Done()
			defer func() { <-limit }()

			for _, tx := range txsByXpub[xPubID] {
				if err = broadcastSyncTransaction(
					ctx, tx,
				); err != nil {
					tx.Client().Logger().Error(ctx,
						fmt.Sprintf("error running broadcast tx for xpub %s, tx %s: %s", xPubID, tx.ID, err.Error()),
					)
					return // stop processing transactions for this xpub if we found an error
				}
			}
		}(xPubID)
	}
	wg.Wait()

	return nil
}

// broadcastSyncTransaction will broadcast transaction related to syncTx record
func broadcastSyncTransaction(ctx context.Context, syncTx *SyncTransaction) error {
	// Successfully capture any panics, convert to readable string and log the error
	defer recoverAndLog(ctx, syncTx.client.Logger())

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, fmt.Sprintf(lockKeyProcessBroadcastTx, syncTx.GetID()), syncTx.Client().Cachestore(),
	)
	defer unlock()
	if err != nil {
		return err
	}

	// Get the transaction HEX
	var txHex string
	if syncTx.transaction != nil && syncTx.transaction.Hex != "" {
		// the transaction has already been retrieved and added to the syncTx object, just use that
		txHex = syncTx.transaction.Hex
	} else {
		// else get hex from DB
		transaction, err := getTransactionByID(
			ctx, "", syncTx.ID, syncTx.GetOptions(false)...,
		)

		if err != nil {
			return err
		}

		if transaction == nil {
			return errors.New("transaction was expected but not found, using ID: " + syncTx.ID)
		}

		txHex = transaction.Hex
	}

	// Broadcast
	var provider string
	if provider, err = syncTx.Client().Chainstate().Broadcast(
		ctx, syncTx.ID, txHex, defaultBroadcastTimeout,
	); err != nil {
		_bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusError, syncActionBroadcast, provider, "broadcast error: "+err.Error(),
		)
		return err
	}

	// Create status message
	message := "broadcast success"

	// Update the sync information
	syncTx.BroadcastStatus = SyncStatusComplete
	syncTx.Results.LastMessage = message
	syncTx.LastAttempt = customTypes.NullTime{
		NullTime: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	}

	syncTx.Results.Results = append(syncTx.Results.Results, &SyncResult{
		Action:        syncActionBroadcast,
		ExecutedAt:    time.Now().UTC(),
		Provider:      provider,
		StatusMessage: message,
	})

	// Update sync status to be ready now
	if syncTx.SyncStatus == SyncStatusPending {
		syncTx.SyncStatus = SyncStatusReady
	}

	// Update the sync transaction record
	if err = syncTx.Save(ctx); err != nil {
		_bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusError, syncActionBroadcast, "internal", err.Error(),
		)
		return err
	}

	// Fire a notification
	notify(notifications.EventTypeBroadcast, syncTx)

	return nil
}

/////////////////

// _syncTxDataFromChain will process the sync transaction record, or save the failure
func _syncTxDataFromChain(ctx context.Context, syncTx *SyncTransaction, transaction *Transaction) error {
	// Successfully capture any panics, convert to readable string and log the error
	defer recoverAndLog(ctx, syncTx.client.Logger())

	var err error

	// Get the transaction
	if transaction == nil {
		if transaction, err = getTransactionByID(
			ctx, "", syncTx.ID, syncTx.GetOptions(false)...,
		); err != nil {
			return err
		}
	}

	if transaction == nil {
		return ErrMissingTransaction
	}

	// Find on-chain
	var txInfo *chainstate.TransactionInfo
	// only mAPI currently provides merkle proof, so QueryTransaction should be used here
	if txInfo, err = syncTx.Client().Chainstate().QueryTransaction(
		ctx, syncTx.ID, chainstate.RequiredOnChain, defaultQueryTxTimeout,
	); err != nil {
		if errors.Is(err, chainstate.ErrTransactionNotFound) {
			syncTx.client.Logger().Info(ctx, fmt.Sprintf("processSyncTransaction(): Transaction %s not found on-chain, will try again later", syncTx.ID))

			_bailAndSaveSyncTransaction(
				ctx, syncTx, SyncStatusReady, syncActionSync, "all", "transaction not found on-chain",
			)
			return nil
		}
		return err
	}

	// validate txInfo
	if txInfo.BlockHash == "" || txInfo.MerkleProof == nil || txInfo.MerkleProof.TxOrID == "" || len(txInfo.MerkleProof.Nodes) == 0 {
		syncTx.client.Logger().Warn(ctx, fmt.Sprintf("processSyncTransaction(): txInfo for %s is invalid, will try again later", syncTx.ID))

		if syncTx.client.IsDebug() {
			txInfoJSON, _ := json.Marshal(txInfo) //nolint:errchkjson // error is not needed
			syncTx.DebugLog(string(txInfoJSON))
		}
		return nil
	}

	transaction.setChainInfo(txInfo)

	// Create status message
	message := "transaction was found on-chain by " + chainstate.ProviderBroadcastClient

	// Save the transaction (should NOT error)
	if err = transaction.Save(ctx); err != nil {
		_bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusError, syncActionSync, "internal", err.Error(),
		)
		return err
	}

	// Update the sync status
	syncTx.SyncStatus = SyncStatusComplete
	syncTx.Results.LastMessage = message
	syncTx.Results.Results = append(syncTx.Results.Results, &SyncResult{
		Action:        syncActionSync,
		ExecutedAt:    time.Now().UTC(),
		Provider:      chainstate.ProviderBroadcastClient,
		StatusMessage: message,
	})

	// Update the sync transaction record
	if err = syncTx.Save(ctx); err != nil {
		_bailAndSaveSyncTransaction(ctx, syncTx, SyncStatusError, syncActionSync, "internal", err.Error())
		return err
	}

	syncTx.client.Logger().Info(ctx, fmt.Sprintf("processSyncTransaction(): Transaction %s processed successfully", syncTx.ID))
	// Done!
	return nil
}

// processP2PTransaction will process the sync transaction record, or save the failure
func processP2PTransaction(ctx context.Context, syncTx *SyncTransaction, transaction *Transaction) error {
	// Successfully capture any panics, convert to readable string and log the error
	defer recoverAndLog(ctx, syncTx.client.Logger())

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, fmt.Sprintf(lockKeyProcessP2PTx, syncTx.GetID()), syncTx.Client().Cachestore(),
	)
	defer unlock()
	if err != nil {
		return err
	}

	// Get the transaction
	if transaction == nil {
		if transaction, err = getTransactionByID(
			ctx, "", syncTx.ID, syncTx.GetOptions(false)...,
		); err != nil {
			return err
		}
	}

	// No draft?
	if len(transaction.DraftID) == 0 {
		_bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusComplete, syncActionP2P, "all", "no draft found, cannot complete p2p",
		)
		return nil
	}

	// Notify any P2P paymail providers associated to the transaction
	var results []*SyncResult
	if results, err = _notifyPaymailProviders(ctx, transaction); err != nil {
		_bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusReady, syncActionP2P, "", err.Error(),
		)
		return err
	}

	// Update if we have some results
	if len(results) > 0 {
		syncTx.Results.Results = append(syncTx.Results.Results, results...)
		syncTx.Results.LastMessage = fmt.Sprintf("notified %d paymail provider(s)", len(results))
	}

	// Save the record
	syncTx.P2PStatus = SyncStatusComplete

	// Update sync status to be ready now
	if syncTx.SyncStatus == SyncStatusPending {
		syncTx.SyncStatus = SyncStatusReady
	}

	if err = syncTx.Save(ctx); err != nil {
		_bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusError, syncActionP2P, "internal", err.Error(),
		)
		return err
	}

	// Done!
	return nil
}

// _notifyPaymailProviders will notify any associated Paymail providers
func _notifyPaymailProviders(ctx context.Context, transaction *Transaction) ([]*SyncResult, error) {
	// First get the draft tx
	draftTx, err := getDraftTransactionID(
		ctx,
		transaction.XPubID,
		transaction.DraftID,
		transaction.GetOptions(false)...,
	)
	if err != nil {
		return nil, err
	} else if draftTx == nil {
		return nil, errors.New("draft not found: " + transaction.DraftID)
	}

	// Loop each output looking for paymail outputs
	var attempts []*SyncResult
	pm := transaction.Client().PaymailClient()
	var payload *paymail.P2PTransactionPayload

	for _, out := range draftTx.Configuration.Outputs {
		if out.PaymailP4 != nil && out.PaymailP4.ResolutionType == ResolutionTypeP2P {

			// Notify each provider with the transaction
			if payload, err = finalizeP2PTransaction(
				ctx,
				pm,
				out.PaymailP4,
				transaction,
			); err != nil {
				return nil, err
			}
			attempts = append(attempts, &SyncResult{
				Action:        syncActionP2P,
				ExecutedAt:    time.Now().UTC(),
				Provider:      out.PaymailP4.ReceiveEndpoint,
				StatusMessage: "success: " + payload.TxID,
			})
		}
	}
	return attempts, nil
}

// utils

func _groupByXpub(scTxs []*SyncTransaction) map[string][]*SyncTransaction {
	txsByXpub := make(map[string][]*SyncTransaction)

	// group transactions by xpub and return including the tx itself
	for _, tx := range scTxs {
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

	return txsByXpub
}

// _bailAndSaveSyncTransaction will save the error message for a sync tx
func _bailAndSaveSyncTransaction(ctx context.Context, syncTx *SyncTransaction, status SyncStatus,
	action, provider, message string,
) {

	if action == syncActionSync {
		syncTx.SyncStatus = status
	} else if action == syncActionP2P {
		syncTx.P2PStatus = status
	} else if action == syncActionBroadcast {
		syncTx.BroadcastStatus = status
	}
	syncTx.LastAttempt = customTypes.NullTime{
		NullTime: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	}
	syncTx.Results.LastMessage = message
	syncTx.Results.Results = append(syncTx.Results.Results, &SyncResult{
		Action:        action,
		ExecutedAt:    time.Now().UTC(),
		Provider:      provider,
		StatusMessage: message,
	})

	if syncTx.IsNew() {
		return // do not save if new record! caller should decide if want to save new record
	}

	_ = syncTx.Save(ctx)
}
