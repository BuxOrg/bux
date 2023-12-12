package bux

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/libsv/go-bt/v2"
	"github.com/mrz1836/go-datastore"
	"github.com/rs/zerolog"
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

func emptyIncomingTx(opts ...ModelOps) *IncomingTransaction {
	return &IncomingTransaction{
		Model:           *NewBaseModel(ModelIncomingTransaction, opts...),
		TransactionBase: TransactionBase{},
		Status:          SyncStatusReady,
	}
}

// newIncomingTransaction will start a new model
func newIncomingTransaction(hex string, opts ...ModelOps) (*IncomingTransaction, error) {
	var btTx *bt.Tx
	var err error

	if btTx, err = bt.NewTxFromString(hex); err != nil {
		return nil, err
	}

	tx := emptyIncomingTx(opts...)
	tx.ID = btTx.TxID()
	tx.Hex = hex
	tx.parsedTx = btTx

	return tx, nil
}

// getIncomingTransactionByID will get the incoming transactions to process
func getIncomingTransactionByID(ctx context.Context, id string, opts ...ModelOps) (*IncomingTransaction, error) {
	// Construct an empty tx
	tx := emptyIncomingTx(opts...)
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
	opts ...ModelOps,
) ([]*IncomingTransaction, error) {
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

func (m *IncomingTransaction) toTransactionDto() *Transaction {
	t := Transaction{}
	t.Hex = m.Hex

	t.parsedTx = m.parsedTx
	t.rawXpubKey = m.rawXpubKey
	t.setXPubID()
	t.setID() //nolint:errcheck,gosec // error is not needed

	t.Metadata = m.Metadata
	t.NumberOfOutputs = uint32(len(m.parsedTx.Outputs))
	t.NumberOfInputs = uint32(len(m.parsedTx.Inputs))

	return &t
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *IncomingTransaction) BeforeCreating(ctx context.Context) error {
	m.Client().Logger().Debug().
		Str("txID", m.GetID()).
		Msgf("starting: %s BeforeCreating hook...", m.Name())

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
	if !m.TransactionBase.hasOneKnownDestination(ctx, m.Client()) {
		return ErrNoMatchingOutputs
	}

	m.Client().Logger().Debug().
		Str("txID", m.GetID()).
		Msgf("end: %s BeforeCreating hook", m.Name())
	return nil
}

// AfterCreated will fire after the model is created
func (m *IncomingTransaction) AfterCreated(_ context.Context) error {
	m.Client().Logger().Debug().
		Str("txID", m.GetID()).
		Msgf("starting: %s AfterCreated hook...", m.Name())

	// todo: this should be refactored into a task
	if err := processIncomingTransaction(context.Background(), m.Client().Logger(), m); err != nil {
		m.Client().Logger().Error().
			Str("txID", m.GetID()).
			Msgf("error processing incoming transaction: %v", err.Error())
	}

	m.Client().Logger().Debug().
		Str("txID", m.GetID()).
		Msgf("end: %s AfterCreated hook", m.Name())
	return nil
}

// Migrate model specific migration on startup
func (m *IncomingTransaction) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableIncomingTransactions), metadataField)
}

// processIncomingTransactions will process incoming transaction records
func processIncomingTransactions(ctx context.Context, logClient *zerolog.Logger, maxTransactions int,
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
		logClient.Info().Msgf("found %d incoming transactions to process", len(records))
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
func processIncomingTransaction(ctx context.Context, logClient *zerolog.Logger,
	incomingTx *IncomingTransaction) error {

	if logClient == nil {
		logClient = incomingTx.client.Logger()
	}

	logClient.Info().Str("txID", incomingTx.GetID()).Msgf("processIncomingTransaction(): transaction: %v", incomingTx)

	// Successfully capture any panics, convert to readable string and log the error
	defer recoverAndLog(incomingTx.client.Logger())

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

		logClient.Error().
			Str("txID", incomingTx.GetID()).
			Msgf("processIncomingTransaction(): error finding transaction %s on chain. Reason: %s", incomingTx.ID, err)

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
			logClient.Info().
				Str("txID", incomingTx.GetID()).
				Msgf("broadcast of transaction was successful using %s. Incoming tx will be processed again.", provider)

			// allow propagation
			time.Sleep(3 * time.Second)
			return nil // reprocess it when triggering the task again
		}

		// Actual error occurred
		bailAndSaveIncomingTransaction(ctx, incomingTx, err.Error())
		return err
	}

	if !txInfo.Valid() {
		logClient.Warn().Str("txID", incomingTx.ID).Msg("txInfo is invalid, will try again later")

		if incomingTx.client.IsDebug() {
			txInfoJSON, _ := json.Marshal(txInfo) //nolint:nolintlint,nilerr,govet,errchkjson // error is not needed
			logClient.Debug().Str("txID", incomingTx.ID).Msg(string(txInfoJSON))
		}
		return nil
	}

	logClient.Info().Str("txID", incomingTx.ID).Msgf("found incoming transaction in %s", txInfo.Provider)

	// Check if we have transaction in DB already
	transaction, _ := getTransactionByID(
		ctx, incomingTx.rawXpubKey, incomingTx.ID, incomingTx.client.DefaultModelOptions()...,
	)

	if transaction == nil {
		// Create the new transaction model
		if transaction, err = newTransactionFromIncomingTransaction(incomingTx); err != nil {
			logClient.Error().Str("txID", incomingTx.ID).Msgf("creating a new tx failed. Reason: %s", err)
			return err
		}

		if err = transaction.processUtxos(ctx); err != nil {
			logClient.Error().
				Str("txID", incomingTx.ID).
				Msgf("processing utxos for tx failed. Reason: %s", err)
			return err
		}
	}

	transaction.setChainInfo(txInfo)

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
