package bux

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/utils"
	"github.com/tonicpow/go-paymail"
)

// SyncTransaction is an object representing the chain-state sync configuration and results for a given transaction
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type SyncTransaction struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	ID              string         `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique transaction id" bson:"_id"`
	Configuration   SyncConfig     `json:"configuration" toml:"configuration" yaml:"configuration" gorm:"<-;type:text;comment:This is the configuration struct in JSON" bson:"configuration"`
	LastAttempt     utils.NullTime `json:"last_attempt" toml:"last_attempt" yaml:"last_attempt" gorm:"<-;comment:When the last broadcast occurred" bson:"last_attempt,omitempty"`
	Results         SyncResults    `json:"results" toml:"results" yaml:"results" gorm:"<-;type:text;comment:This is the results struct in JSON" bson:"results"`
	BroadcastStatus SyncStatus     `json:"broadcast_status" toml:"broadcast_status" yaml:"broadcast_status" gorm:"<-;type:varchar(10);index;comment:This is the status of the broadcast" bson:"broadcast_status"`
	SyncStatus      SyncStatus     `json:"sync_status" toml:"sync_status" yaml:"sync_status" gorm:"<-;type:varchar(10);index;comment:This is the status of the on-chain sync" bson:"sync_status"`
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

	// Sync
	ss := SyncStatusPending
	if !config.Broadcast {
		ss = SyncStatusSkipped
	}

	return &SyncTransaction{
		BroadcastStatus: bs,
		Configuration:   *config,
		ID:              txID,
		Model:           *NewBaseModel(ModelSyncTransaction, opts...),
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
func getSyncTransactionsByConditions(ctx context.Context, conditions map[string]interface{}, queryParams *datastore.QueryParams,
	opts ...ModelOps) ([]*SyncTransaction, error) {

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

// isSkipped will return true if Broadcasting & SyncOnChain are both skipped
func (m *SyncTransaction) isSkipped() bool {
	return m.BroadcastStatus == SyncStatusSkipped && m.SyncStatus == SyncStatusSkipped
}

// GetModelName will get the name of the current model
func (m *SyncTransaction) GetModelName() string {
	return ModelSyncTransaction.String()
}

// GetModelTableName will get the db table name of the current model
func (m *SyncTransaction) GetModelTableName() string {
	return tableSyncTransactions
}

// Save will Save the model into the Datastore
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

	// Set status
	m.BroadcastStatus = SyncStatusReady
	m.SyncStatus = SyncStatusPending

	// Make sure ID is valid
	if len(m.ID) == 0 {
		return ErrMissingFieldID
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
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
	syncTask := m.Name() + "_sync"
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
	broadcastTask := m.Name() + "_broadcast"

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
	return tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(broadcastTask),
		TaskName:       broadcastTask,
	})
}

// Migrate model specific migration on startup
func (m *SyncTransaction) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableSyncTransactions), metadataField)
}

// processSyncTransactions will process sync transaction records
func processSyncTransactions(ctx context.Context, maxTransactions int, opts ...ModelOps) error {

	queryParams := &datastore.QueryParams{Page: 1, PageSize: maxTransactions}

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
			ctx, records[index],
		); err != nil {
			return err
		}
	}

	return nil
}

// processBroadcastTransactions will process sync transaction records
func processBroadcastTransactions(ctx context.Context, maxTransactions int, opts ...ModelOps) error {

	queryParams := &datastore.QueryParams{Page: 1, PageSize: maxTransactions}

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

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, fmt.Sprintf(lockKeyProcessSyncTx, syncTx.GetID()), syncTx.Client().Cachestore(),
	)
	defer unlock()
	if err != nil {
		return err
	}

	// Get the transaction
	var transaction *Transaction
	if transaction, err = getTransactionByID(
		ctx, syncTx.rawXpubKey, syncTx.ID, syncTx.GetOptions(false)...,
	); err != nil {
		return err
	}

	// Broadcast
	var provider string
	if provider, err = syncTx.Client().Chainstate().Broadcast(
		ctx, syncTx.ID, transaction.Hex, defaultBroadcastTimeout,
	); err != nil {
		bailAndSaveSyncTransaction(ctx, syncTx, SyncStatusError, provider, "broadcast error: "+err.Error())
		return nil // nolint: nilerr // error is not needed
	}

	// Create status message
	message := "transaction was broadcasted"

	// Update the sync status
	syncTx.BroadcastStatus = SyncStatusComplete
	syncTx.Results.LastMessage = message
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
		bailAndSaveSyncTransaction(ctx, syncTx, SyncStatusError, "internal", err.Error())
		return err
	}

	// Notify any P2P paymail providers associated to the transaction
	if len(transaction.DraftID) > 0 {
		var results []*SyncResult
		if results, err = notifyPaymailProviders(transaction); err != nil {
			return err
		}
		if len(results) > 0 {
			syncTx.Results.Results = append(syncTx.Results.Results, results...)
			syncTx.Results.LastMessage = fmt.Sprintf("notified %d paymail provider(s)", len(results))
			if err = syncTx.Save(ctx); err != nil {
				return err
			}
		}
	}

	// Done!
	return nil
}

// processSyncTransaction will process the sync transaction record, or save the failure
func processSyncTransaction(ctx context.Context, syncTx *SyncTransaction) error {

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
			bailAndSaveSyncTransaction(ctx, syncTx, SyncStatusReady, "all", "transaction not found on-chain")
			return nil
		}
		return err
	}

	// Get the transaction
	var transaction *Transaction
	if transaction, err = getTransactionByID(
		ctx, "", txInfo.ID, syncTx.GetOptions(false)...,
	); err != nil {
		return err
	}

	// Add additional information (if found on-chain)
	transaction.BlockHash = txInfo.BlockHash
	transaction.BlockHeight = uint64(txInfo.BlockHeight)

	// Create status message
	message := "transaction was found on-chain by " + txInfo.Provider

	// Save the transaction (should NOT error)
	if err = transaction.Save(ctx); err != nil {
		bailAndSaveSyncTransaction(ctx, syncTx, SyncStatusError, "internal", err.Error())
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
		bailAndSaveSyncTransaction(ctx, syncTx, SyncStatusError, "internal", err.Error())
		return err
	}

	// Done!
	return nil
}

// bailAndSaveSyncTransaction try to save the error message
func bailAndSaveSyncTransaction(ctx context.Context, syncTx *SyncTransaction, status SyncStatus, provider, message string) {
	syncTx.SyncStatus = status
	syncTx.LastAttempt = utils.NullTime{
		NullTime: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	}
	syncTx.Results.LastMessage = message
	syncTx.Results.Results = append(syncTx.Results.Results, &SyncResult{
		Action:        syncActionSync,
		ExecutedAt:    time.Now().UTC(),
		Provider:      provider,
		StatusMessage: message,
	})
	_ = syncTx.Save(ctx)
}

// notifyPaymailProviders will notify any associated Paymail providers
func notifyPaymailProviders(transaction *Transaction) ([]*SyncResult, error) {

	// First get the draft tx
	draftTx, err := getDraftTransactionID(
		context.Background(),
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

	transaction.Client().Logger().Info(
		context.Background(), fmt.Sprintf("outputs found: %d", len(draftTx.Configuration.Outputs)),
	)

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
