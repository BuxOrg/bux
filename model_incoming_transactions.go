package bux

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/libsv/go-bt/v2"
	"github.com/mrz1836/go-datastore"
	zLogger "github.com/mrz1836/go-logger"
)

// IncomingTransaction is an object representing the incoming (external) transaction (for pre-processing)
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type IncomingTransaction struct {
	// Base model
	Model `bson:",inline"`

	// Standard transaction model base fields
	TransactionBase `bson:",inline"`

	// Model specific fields
	Status        SyncStatus `json:"status" toml:"status" yaml:"status" gorm:"<-;type:varchar(10);index;comment:This is the status of processing the transaction" bson:"status"`
	StatusMessage string     `json:"status_message" toml:"status_message" yaml:"status_message" gorm:"<-;type:varchar(512);comment:This is the status message or error" bson:"status_message"`
}

// newIncomingTransaction will start a new model
func newIncomingTransaction(txID, hex string, opts ...ModelOps) (tx *IncomingTransaction) {

	// Create the model
	tx = &IncomingTransaction{
		Model: *NewBaseModel(ModelIncomingTransaction, opts...),
		TransactionBase: TransactionBase{
			ID:  txID,
			Hex: hex,
		},
		Status: SyncStatusReady,
	}

	// Attempt to parse
	if len(hex) > 0 {
		tx.TransactionBase.parsedTx, _ = bt.NewTxFromString(hex)
	}

	return
}

// getIncomingTransactionByID will get the incoming transactions to process
func getIncomingTransactionByID(ctx context.Context, id string, opts ...ModelOps) (*IncomingTransaction, error) {
	// Construct an empty tx
	tx := newIncomingTransaction("", "", opts...)
	tx.ID = id

	// Get the record
	if err := Get(ctx, tx, nil, false, defaultDatabaseReadTimeout, false); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return tx, nil
}

// getIncomingTransactionsToProcess will get the incoming transactions to process
func getIncomingTransactionsToProcess(ctx context.Context, queryParams *datastore.QueryParams,
	opts ...ModelOps) ([]*IncomingTransaction, error) {

	// Construct an empty model
	var models []IncomingTransaction
	conditions := map[string]interface{}{
		statusField: statusReady,
	}

	if queryParams == nil {
		queryParams = &datastore.QueryParams{
			Page:     0,
			PageSize: 0,
		}
	}
	queryParams.OrderByField = idField
	queryParams.SortDirection = datastore.SortAsc

	// Get the record
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
	txs := make([]*IncomingTransaction, 0)
	for index := range models {
		models[index].enrich(ModelIncomingTransaction, opts...)
		txs = append(txs, &models[index])
	}

	return txs, nil
}

// GetModelName will get the name of the current model
func (m *IncomingTransaction) GetModelName() string {
	return ModelIncomingTransaction.String()
}

// GetModelTableName will get the db table name of the current model
func (m *IncomingTransaction) GetModelTableName() string {
	return tableIncomingTransactions
}

// Save will save the model into the Datastore
func (m *IncomingTransaction) Save(ctx context.Context) error {
	return Save(ctx, m)
}

// GetID will get the ID
func (m *IncomingTransaction) GetID() string {
	return m.ID
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *IncomingTransaction) BeforeCreating(ctx context.Context) error {
	m.DebugLog("starting: [" + m.name.String() + "] BeforeCreating hook...")

	// Set status
	m.Status = SyncStatusReady

	// Make sure ID is valid
	if len(m.ID) == 0 {
		return ErrMissingFieldID
	}
	if len(m.Hex) == 0 {
		return ErrMissingFieldHex
	}

	// Attempt to parse
	if len(m.Hex) > 0 && m.TransactionBase.parsedTx == nil {
		m.TransactionBase.parsedTx, _ = bt.NewTxFromString(m.Hex)
	}

	// Require the tx to be parsed
	if m.TransactionBase.parsedTx == nil {
		return ErrTransactionNotParsed
	}

	// Check that the transaction has >= 1 known destination
	if !m.TransactionBase.hasOneKnownDestination(ctx, m.Client(), m.GetOptions(false)...) {
		return ErrNoMatchingOutputs
	}

	// Match a known destination
	// todo: this can be optimized searching X records at a time vs loop->query->loop->query
	matchingOutput := false
	lockingScript := ""
	opts := m.GetOptions(false)
	for index := range m.TransactionBase.parsedTx.Outputs {
		lockingScript = m.TransactionBase.parsedTx.Outputs[index].LockingScript.String()
		destination, err := getDestinationWithCache(ctx, m.Client(), "", "", lockingScript, opts...)
		if err != nil {
			m.Client().Logger().Warn(ctx, "error getting destination: "+err.Error())
		} else if destination != nil && destination.LockingScript == lockingScript {
			matchingOutput = true
			break
		}
	}

	// Does not match any known destination
	if !matchingOutput {
		return ErrNoMatchingOutputs
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// AfterCreated will fire after the model is created
func (m *IncomingTransaction) AfterCreated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	// todo: this should be refactored into a task
	// go func(incomingTx *IncomingTransaction) {
	if err := processIncomingTransaction(context.Background(), nil, m); err != nil {
		m.Client().Logger().Error(ctx, "error processing incoming transaction: "+err.Error())
	}
	// }(m)

	m.DebugLog("end: " + m.Name() + " AfterCreated hook...")
	return nil
}

// Migrate model specific migration on startup
func (m *IncomingTransaction) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableIncomingTransactions), metadataField)
}

// RegisterTasks will register the model specific tasks on client initialization
func (m *IncomingTransaction) RegisterTasks() error {

	// No task manager loaded?
	tm := m.Client().Taskmanager()
	if tm == nil {
		return nil
	}

	// Register the task locally (cron task - set the defaults)
	processTask := m.Name() + "_process"
	ctx := context.Background()

	// Register the task
	if err := tm.RegisterTask(&taskmanager.Task{
		Name:       processTask,
		RetryLimit: 1,
		Handler: func(client ClientInterface) error {
			if taskErr := taskProcessIncomingTransactions(ctx, client.Logger(), WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+processTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	// Run the task periodically
	return tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(processTask),
		TaskName:       processTask,
	})
}

// processIncomingTransactions will process incoming transaction records
func processIncomingTransactions(ctx context.Context, logClient zLogger.GormLoggerInterface, maxTransactions int,
	opts ...ModelOps) error {

	queryParams := &datastore.QueryParams{Page: 1, PageSize: maxTransactions}

	// Get x records:
	records, err := getIncomingTransactionsToProcess(
		ctx, queryParams, opts...,
	)
	if err != nil {
		return err
	} else if len(records) == 0 {
		return nil
	}

	if logClient != nil {
		logClient.Info(ctx, fmt.Sprintf("found %d incoming transactions to process", len(records)))
	}

	// Process the incoming transaction
	for index := range records {
		if err = processIncomingTransaction(
			ctx, logClient, records[index],
		); err != nil {
			return err
		}
	}

	return nil
}

// processIncomingTransaction will process the incoming transaction record into a transaction, or save the failure
func processIncomingTransaction(ctx context.Context, logClient zLogger.GormLoggerInterface,
	incomingTx *IncomingTransaction) error {

	if logClient != nil {
		logClient.Info(ctx, fmt.Sprintf("processing incoming transaction: %v", incomingTx))
	}

	// Successfully capture any panics, convert to readable string and log the error
	defer func() {
		if err := recover(); err != nil {
			incomingTx.Client().Logger().Error(ctx,
				fmt.Sprintf(
					"panic: %v - stack trace: %v", err,
					strings.ReplaceAll(string(debug.Stack()), "\n", ""),
				),
			)
		}
	}()

	// Create the lock and set the release for after the function completes
	unlock, err := newWriteLock(
		ctx, fmt.Sprintf(lockKeyProcessIncomingTx, incomingTx.GetID()), incomingTx.Client().Cachestore(),
	)
	defer unlock()
	if err != nil {
		return err
	}

	// Find in mempool or on-chain
	var txInfo *chainstate.TransactionInfo
	if txInfo, err = incomingTx.Client().Chainstate().QueryTransactionFastest(
		ctx, incomingTx.ID, chainstate.RequiredInMempool, defaultQueryTxTimeout,
	); err != nil {

		if logClient != nil {
			logClient.Error(ctx, fmt.Sprintf("error finding transaction %s on chain: %s", incomingTx.ID, err.Error()))
		}

		// TX might not have been broadcast yet? (race condition, or it was never broadcast...)
		if errors.Is(err, chainstate.ErrTransactionNotFound) {
			var provider string

			// Broadcast and detect if there is a real error
			if provider, err = incomingTx.Client().Chainstate().Broadcast(
				ctx, incomingTx.ID, incomingTx.Hex, defaultQueryTxTimeout,
			); err != nil {
				bailAndSaveIncomingTransaction(ctx, incomingTx, "tx was not found using all providers, attempted broadcast, "+err.Error())
				return err
			}

			// Broadcast was successful, so the transaction was accepted by the network, continue processing like before
			if logClient != nil {
				logClient.Info(ctx, fmt.Sprintf("broadcast of transaction was successful using %s", provider))
			}
			// allow propagation
			time.Sleep(3 * time.Second)
			if txInfo, err = incomingTx.Client().Chainstate().QueryTransactionFastest(
				ctx, incomingTx.ID, chainstate.RequiredInMempool, defaultQueryTxTimeout,
			); err != nil {
				incomingTx.Status = statusReady
				incomingTx.StatusMessage = "tx was not found on-chain, attempting to broadcast using provider: " + provider
				_ = incomingTx.Save(ctx)
				return err
			}
		} else {
			// Actual error occurred
			bailAndSaveIncomingTransaction(ctx, incomingTx, err.Error())
			return err
		}
	}

	if logClient != nil {
		logClient.Info(ctx, fmt.Sprintf("found incoming transaction %s in %s", incomingTx.ID, txInfo.Provider))
	}

	// Create the new transaction model
	transaction := newTransactionFromIncomingTransaction(incomingTx)

	// Get the transaction by ID
	if tx, _ := getTransactionByID(
		ctx, transaction.rawXpubKey, transaction.TransactionBase.ID, transaction.client.DefaultModelOptions()...,
	); tx != nil {
		transaction = tx
	}
	// Add additional information (if found on-chain)
	transaction.BlockHeight = uint64(txInfo.BlockHeight)
	transaction.BlockHash = txInfo.BlockHash

	// Create status message
	onChain := len(transaction.BlockHash) > 0 || transaction.BlockHeight > 0
	message := "transaction was found in mempool by " + txInfo.Provider
	if onChain {
		message = "transaction was found on-chain by " + txInfo.Provider
	}

	// Save (add) the transaction (should NOT error)
	if err = transaction.Save(ctx); err != nil {
		bailAndSaveIncomingTransaction(ctx, incomingTx, err.Error())
		return err
	}

	// Update (or delete?) the incoming transaction record
	incomingTx.Status = statusComplete
	incomingTx.StatusMessage = message
	if err = incomingTx.Save(ctx); err != nil {
		bailAndSaveIncomingTransaction(ctx, incomingTx, err.Error())
		return err
	}

	// Done!
	return nil
}

// bailAndSaveIncomingTransaction try to save the error message
func bailAndSaveIncomingTransaction(ctx context.Context, incomingTx *IncomingTransaction, errorMessage string) {
	incomingTx.Status = statusError
	incomingTx.StatusMessage = errorMessage
	_ = incomingTx.Save(ctx)
}
