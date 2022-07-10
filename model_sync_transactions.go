package bux

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/notifications"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/mrz1836/go-datastore"
	customTypes "github.com/mrz1836/go-datastore/custom_types"
	"github.com/tonicpow/go-paymail"
)

// SyncTransaction is an object representing the chain-state sync configuration and results for a given transaction
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type SyncTransaction struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	ID              string               `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique transaction id" bson:"_id"`
	Configuration   SyncConfig           `json:"configuration" toml:"configuration" yaml:"configuration" gorm:"<-;type:text;comment:This is the configuration struct in JSON" bson:"configuration"`
	LastAttempt     customTypes.NullTime `json:"last_attempt" toml:"last_attempt" yaml:"last_attempt" gorm:"<-;comment:When the last broadcast occurred" bson:"last_attempt,omitempty"`
	Results         SyncResults          `json:"results" toml:"results" yaml:"results" gorm:"<-;type:text;comment:This is the results struct in JSON" bson:"results"`
	BroadcastStatus SyncStatus           `json:"broadcast_status" toml:"broadcast_status" yaml:"broadcast_status" gorm:"<-;type:varchar(10);index;comment:This is the status of the broadcast" bson:"broadcast_status"`
	P2PStatus       SyncStatus           `json:"p2p_status" toml:"p2p_status" yaml:"p2p_status" gorm:"<-;column:p2p_status;type:varchar(10);index;comment:This is the status of the p2p paymail requests" bson:"p2p_status"`
	SyncStatus      SyncStatus           `json:"sync_status" toml:"sync_status" yaml:"sync_status" gorm:"<-;type:varchar(10);index;comment:This is the status of the on-chain sync" bson:"sync_status"`
}

// newSyncTransaction will start a new model (config is required)
func newSyncTransaction(txID string, config *SyncConfig, opts ...ModelOps) *SyncTransaction {

	// Do not allow making a model without the configuration
	if config == nil {
		return nil
	}

	// Broadcasting
	bs := SyncStatusReady
	if !config.Broadcast {
		bs = SyncStatusSkipped
	}

	// Notify Paymail P2P
	ps := SyncStatusPending
	if !config.PaymailP2P {
		ps = SyncStatusSkipped
	}

	// Sync
	ss := SyncStatusPending
	if !config.SyncOnChain {
		ss = SyncStatusSkipped
	}

	return &SyncTransaction{
		BroadcastStatus: bs,
		Configuration:   *config,
		ID:              txID,
		Model:           *NewBaseModel(ModelSyncTransaction, opts...),
		P2PStatus:       ps,
		SyncStatus:      ss,
	}
}

// getTransactionsToBroadcast will get the sync transactions to broadcast
func getTransactionsToBroadcast(ctx context.Context, queryParams *datastore.QueryParams,
	opts ...ModelOps) ([]*SyncTransaction, error) {

	// Get the records by status
	txs, err := getSyncTransactionsByConditions(
		ctx,
		map[string]interface{}{
			broadcastStatusField: SyncStatusReady.String(),
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
	opts ...ModelOps) ([]*SyncTransaction, error) {

	// Get the records by status
	txs, err := getSyncTransactionsByConditions(
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

// getTransactionsToSync will get the sync transactions to sync
func getTransactionsToSync(ctx context.Context, queryParams *datastore.QueryParams,
	opts ...ModelOps) ([]*SyncTransaction, error) {

	// Get the records by status
	txs, err := getSyncTransactionsByConditions(
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

// getTransactionsToSync will get the sync transactions to sync
func getSyncTransactionsByConditions(ctx context.Context, conditions map[string]interface{},
	queryParams *datastore.QueryParams, opts ...ModelOps) ([]*SyncTransaction, error) {

	if queryParams == nil {
		queryParams = &datastore.QueryParams{}
	}
	queryParams.OrderByField = idField
	queryParams.SortDirection = datastore.SortAsc

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

// isSkipped will return true if Broadcasting, P2P and SyncOnChain are all skipped
func (m *SyncTransaction) isSkipped() bool {
	return m.BroadcastStatus == SyncStatusSkipped &&
		m.SyncStatus == SyncStatusSkipped &&
		m.P2PStatus == SyncStatusSkipped
}

// GetModelName will get the name of the current model
func (m *SyncTransaction) GetModelName() string {
	return ModelSyncTransaction.String()
}

// GetModelTableName will get the db table name of the current model
func (m *SyncTransaction) GetModelTableName() string {
	return tableSyncTransactions
}

// Save will save the model into the Datastore
func (m *SyncTransaction) Save(ctx context.Context) error {
	return Save(ctx, m)
}

// GetID will get the ID
func (m *SyncTransaction) GetID() string {
	return m.ID
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *SyncTransaction) BeforeCreating(_ context.Context) error {
	m.DebugLog("starting: [" + m.name.String() + "] BeforeCreating hook...")

	// Make sure ID is valid
	if len(m.ID) == 0 {
		return ErrMissingFieldID
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *SyncTransaction) AfterCreated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	// Should we broadcast immediately?
	if m.Configuration.Broadcast &&
		m.Configuration.BroadcastInstant {
		if err := processBroadcastTransaction(
			ctx, m,
		); err != nil {
			// return err (do not return and fail the record creation)
			m.Client().Logger().Error(ctx, "error running broadcast tx: "+err.Error())
		}
	}

	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

// RegisterTasks will register the model specific tasks on client initialization
func (m *SyncTransaction) RegisterTasks() error {

	// No task manager loaded?
	tm := m.Client().Taskmanager()
	if tm == nil {
		return nil
	}

	// Register the task locally (cron task - set the defaults)
	syncTask := m.Name() + "_" + syncActionSync
	ctx := context.Background()

	// Register the task
	if err := tm.RegisterTask(&taskmanager.Task{
		Name:       syncTask,
		RetryLimit: 1,
		Handler: func(client ClientInterface) error {
			if taskErr := taskSyncTransactions(ctx, client.Logger(), WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+syncTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	// Run the task periodically
	err := tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(syncTask),
		TaskName:       syncTask,
	})
	if err != nil {
		return err
	}

	// Register the task locally (cron task - set the defaults)
	broadcastTask := m.Name() + "_" + syncActionBroadcast

	// Register the task
	if err = tm.RegisterTask(&taskmanager.Task{
		Name:       broadcastTask,
		RetryLimit: 1,
		Handler: func(client ClientInterface) error {
			if taskErr := taskBroadcastTransactions(ctx, client.Logger(), WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+broadcastTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	// Run the task periodically
	if err = tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(broadcastTask),
		TaskName:       broadcastTask,
	}); err != nil {
		return err
	}

	// Register the task locally (cron task - set the defaults)
	p2pTask := m.Name() + "_" + syncActionP2P

	// Register the task
	if err = tm.RegisterTask(&taskmanager.Task{
		Name:       p2pTask,
		RetryLimit: 1,
		Handler: func(client ClientInterface) error {
			if taskErr := taskNotifyP2P(ctx, client.Logger(), WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+p2pTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	// Run the task periodically
	return tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(p2pTask),
		TaskName:       p2pTask,
	})
}

// Migrate model specific migration on startup
func (m *SyncTransaction) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableSyncTransactions), metadataField)
}

// processSyncTransactions will process sync transaction records
func processSyncTransactions(ctx context.Context, maxTransactions int, opts ...ModelOps) error {

	queryParams := &datastore.QueryParams{
		Page:          1,
		PageSize:      maxTransactions,
		OrderByField:  "created_at",
		SortDirection: "asc",
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
		if err = processSyncTransaction(
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
		OrderByField:  "created_at",
		SortDirection: "asc",
	}

	// Get x records
	records, err := getTransactionsToBroadcast(
		ctx, queryParams, opts...,
	)
	if err != nil {
		return err
	} else if len(records) == 0 {
		return nil
	}

	// Process the incoming transaction
	for index := range records {
		if err = processBroadcastTransaction(
			ctx, records[index],
		); err != nil {
			return err
		}
	}

	return nil
}

// processBroadcastTransaction will process the sync transaction record, or save the failure
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

	// Broadcast
	var provider string
	if provider, err = syncTx.Client().Chainstate().Broadcast(
		ctx, syncTx.ID, txHex, defaultBroadcastTimeout,
	); err != nil {
		bailAndSaveSyncTransaction(
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
		bailAndSaveSyncTransaction(
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
			return processP2PTransaction(ctx, syncTx, transaction)
		} else if syncTx.P2PStatus == SyncStatusSkipped && syncTx.SyncStatus == SyncStatusReady {
			return processSyncTransaction(ctx, syncTx, transaction)
		}
	}
	return nil
}

// processSyncTransaction will process the sync transaction record, or save the failure
func processSyncTransaction(ctx context.Context, syncTx *SyncTransaction, transaction *Transaction) error {

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

	// Find on-chain
	var txInfo *chainstate.TransactionInfo
	if txInfo, err = syncTx.Client().Chainstate().QueryTransactionFastest(
		ctx, syncTx.ID, chainstate.RequiredOnChain, defaultQueryTxTimeout,
	); err != nil {
		if errors.Is(err, chainstate.ErrTransactionNotFound) {
			bailAndSaveSyncTransaction(
				ctx, syncTx, SyncStatusReady, syncActionSync, "all", "transaction not found on-chain",
			)
			return nil
		}
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

	// Add additional information (if found on-chain)
	transaction.BlockHash = txInfo.BlockHash
	transaction.BlockHeight = uint64(txInfo.BlockHeight)

	// Create status message
	message := "transaction was found on-chain by " + txInfo.Provider

	// Save the transaction (should NOT error)
	if err = transaction.Save(ctx); err != nil {
		bailAndSaveSyncTransaction(
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
		Provider:      txInfo.Provider,
		StatusMessage: message,
	})

	// Update the sync transaction record
	if err = syncTx.Save(ctx); err != nil {
		bailAndSaveSyncTransaction(ctx, syncTx, SyncStatusError, syncActionSync, "internal", err.Error())
		return err
	}

	// Done!
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
		if err = processP2PTransaction(
			ctx, records[index], nil,
		); err != nil {
			return err
		}
	}

	return nil
}

// processP2PTransaction will process the sync transaction record, or save the failure
func processP2PTransaction(ctx context.Context, syncTx *SyncTransaction, transaction *Transaction) error {

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
		bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusComplete, syncActionP2P, "all", "no draft found, cannot complete p2p",
		)
		return nil
	}

	// Notify any P2P paymail providers associated to the transaction
	var results []*SyncResult
	if results, err = notifyPaymailProviders(ctx, transaction); err != nil {
		bailAndSaveSyncTransaction(
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
		bailAndSaveSyncTransaction(
			ctx, syncTx, SyncStatusError, syncActionP2P, "internal", err.Error(),
		)
		return err
	}

	// Done!
	return nil
}

// bailAndSaveSyncTransaction will save the error message for a sync tx
func bailAndSaveSyncTransaction(ctx context.Context, syncTx *SyncTransaction, status SyncStatus,
	action, provider, message string) {
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

// notifyPaymailProviders will notify any associated Paymail providers
func notifyPaymailProviders(ctx context.Context, transaction *Transaction) ([]*SyncResult, error) {

	// First get the draft tx
	draftTx, err := getDraftTransactionID(
		ctx,
		transaction.xPubID,
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
				pm,
				out.PaymailP4.Alias,
				out.PaymailP4.Domain,
				out.PaymailP4.ReceiveEndpoint,
				out.PaymailP4.ReferenceID,
				out.PaymailP4.Note,
				out.PaymailP4.FromPaymail,
				transaction.Hex,
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
