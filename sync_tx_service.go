package bux

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
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
		if err = _processSyncTransaction(
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
	txsByXpub, err := getTransactionsToBroadcast(
		ctx, queryParams, opts...,
	)
	if err != nil {
		return err
	} else if len(txsByXpub) == 0 {
		return nil
	}

	wg := new(sync.WaitGroup)
	// we limit the number of concurrent broadcasts to the number of cpus*2, since there is lots of IO wait
	limit := make(chan bool, runtime.NumCPU()*2)

	// Process the transactions per xpub, in parallel
	for xPubID := range txsByXpub {
		limit <- true // limit the number of routines running at the same time
		wg.Add(1)
		go func(xPubID string) {
			defer wg.Done()
			defer func() { <-limit }()

			for _, tx := range txsByXpub[xPubID] {
				if err = processBroadcastTransaction(
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

// processP2PTransactions will process transactions for p2p notifications
func processP2PTransactions(ctx context.Context, maxTransactions int, opts ...ModelOps) error {
	queryParams := &datastore.QueryParams{
		Page:          1,
		PageSize:      maxTransactions,
		OrderByField:  "created_at",
		SortDirection: "asc",
	}

	// Get x records
	records, err := getTransactionsToNotifyP2P(
		ctx, queryParams, opts...,
	)
	if err != nil {
		return err
	} else if len(records) == 0 {
		return nil
	}

	// Process the incoming transaction
	for index := range records {
		if err = _processP2PTransaction(
			ctx, records[index], nil,
		); err != nil {
			return err
		}
	}

	return nil
}

// processBroadcastTransaction will process a sync transaction record and broadcast it
func processBroadcastTransaction(ctx context.Context, syncTx *SyncTransaction) error {
	// Successfully capture any panics, convert to readable string and log the error
	defer func() {
		if err := recover(); err != nil {
			syncTx.Client().Logger().Error(ctx,
				fmt.Sprintf(
					"panic: %v - stack trace: %v", err,
					strings.ReplaceAll(string(debug.Stack()), "\n", ""),
				),
			)
		}
	}()

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, fmt.Sprintf(lockKeyProcessBroadcastTx, syncTx.GetID()), syncTx.Client().Cachestore(),
	)
	defer unlock()
	if err != nil {
		return err
	}

	// Get the transaction
	var transaction *Transaction
	var incomingTransaction *IncomingTransaction
	var txHex string
	if syncTx.transaction != nil && syncTx.transaction.Hex != "" {
		// the transaction has already been retrieved and added to the syncTx object, just use that
		transaction = syncTx.transaction
		txHex = transaction.Hex
	} else {
		if transaction, err = getTransactionByID(
			ctx, "", syncTx.ID, syncTx.GetOptions(false)...,
		); err != nil {
			return err
		} else if transaction == nil {
			// maybe this is only an incoming transaction, let's try to find that and broadcast
			// the processing of incoming transactions should then pick it up in the next job run
			if incomingTransaction, err = getIncomingTransactionByID(ctx, syncTx.ID, syncTx.GetOptions(false)...); err != nil {
				return err
			} else if incomingTransaction == nil {
				return errors.New("transaction was expected but not found, using ID: " + syncTx.ID)
			}
			txHex = incomingTransaction.Hex
		} else {
			txHex = transaction.Hex
		}
	}

	// Broadcast
	var provider string
	if provider, err = syncTx.Client().Chainstate().Broadcast(
		ctx, syncTx.ID, txHex, defaultBroadcastTimeout,
	); err != nil {
		_bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusError, syncActionBroadcast, provider, "broadcast error: "+err.Error(),
		)
		return nil //nolint:nolintlint,nilerr // error is not needed
	}

	// Create status message
	message := "broadcast success"

	// process the incoming transaction before finishing the sync
	if incomingTransaction != nil {
		// give the transaction some time to propagate through the network
		time.Sleep(3 * time.Second)

		// we don't need to handle the error here, this is only to speed up the processing
		// job will pick it up later if needed
		if err = processIncomingTransaction(ctx, nil, incomingTransaction); err == nil {
			// again ignore the error, if an error occurs the transaction will be set and will be processed further
			transaction, _ = getTransactionByID(ctx, "", syncTx.ID, syncTx.GetOptions(false)...)
		}
	}

	// Update the sync information
	syncTx.BroadcastStatus = SyncStatusComplete
	syncTx.Results.LastMessage = message
	syncTx.LastAttempt = customTypes.NullTime{
		NullTime: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	}

	// Trim the results to the last 20
	if len(syncTx.Results.Results) >= 19 {
		syncTx.Results.Results = syncTx.Results.Results[1:]
	}

	syncTx.Results.Results = append(syncTx.Results.Results, &SyncResult{
		Action:        syncActionBroadcast,
		ExecutedAt:    time.Now().UTC(),
		Provider:      provider,
		StatusMessage: message,
	})

	// Update the P2P status
	if syncTx.P2PStatus == SyncStatusPending {
		syncTx.P2PStatus = SyncStatusReady
	}

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

	// Notify any P2P paymail providers associated to the transaction
	// but only if we actually found the transaction in the transactions' collection, otherwise this was an incoming
	// transaction that needed to be broadcast and was not successfully processed after the broadcast
	if transaction != nil {
		if syncTx.P2PStatus == SyncStatusReady {
			return _processP2PTransaction(ctx, syncTx, transaction)
		} else if syncTx.P2PStatus == SyncStatusSkipped && syncTx.SyncStatus == SyncStatusReady {
			return _processSyncTransaction(ctx, syncTx, transaction)
		}
	}
	return nil
}

/////////////////

// _processSyncTransaction will process the sync transaction record, or save the failure
func _processSyncTransaction(ctx context.Context, syncTx *SyncTransaction, transaction *Transaction) error {
	// Successfully capture any panics, convert to readable string and log the error
	defer func() {
		if err := recover(); err != nil {
			syncTx.Client().Logger().Error(ctx,
				fmt.Sprintf(
					"panic: %v - stack trace: %v", err,
					strings.ReplaceAll(string(debug.Stack()), "\n", ""),
				),
			)
		}
	}()

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, fmt.Sprintf(lockKeyProcessSyncTx, syncTx.GetID()), syncTx.Client().Cachestore(),
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
			txInfoJSON, _ := json.Marshal(txInfo) //nolint:nolintlint,nilerr // error is not needed
			syncTx.DebugLog(string(txInfoJSON))
		}
		return nil
	}

	// Add additional information (if found on-chain)
	transaction.BlockHash = txInfo.BlockHash
	transaction.BlockHeight = uint64(txInfo.BlockHeight)
	transaction.MerkleProof = MerkleProof(*txInfo.MerkleProof)

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

// _processP2PTransaction will process the sync transaction record, or save the failure
func _processP2PTransaction(ctx context.Context, syncTx *SyncTransaction, transaction *Transaction) error {
	// Successfully capture any panics, convert to readable string and log the error
	defer func() {
		if err := recover(); err != nil {
			syncTx.Client().Logger().Error(ctx,
				fmt.Sprintf(
					"panic: %v - stack trace: %v", err,
					strings.ReplaceAll(string(debug.Stack()), "\n", ""),
				),
			)
		}
	}()

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
	_ = syncTx.Save(ctx)
}