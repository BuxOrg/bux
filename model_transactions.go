package bux

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/BuxOrg/bux/notifications"
	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
	"github.com/mrz1836/go-datastore"
)

// TransactionBase is the same fields share between multiple transaction models
type TransactionBase struct {
	ID  string `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique id (hash of the transaction hex)" bson:"_id"`
	Hex string `json:"hex" toml:"hex" yaml:"hex" gorm:"<-:create;type:text;comment:This is the raw transaction hex" bson:"hex"`

	// Private for internal use
	parsedTx *bt.Tx `gorm:"-" bson:"-"` // The go-bt version of the transaction
}

// TransactionDirection String describing the direction of the transaction (in / out)
type TransactionDirection string

const (
	// TransactionDirectionIn The transaction is coming in to the wallet of the xpub
	TransactionDirectionIn TransactionDirection = "incoming"

	// TransactionDirectionOut The transaction is going out of to the wallet of the xpub
	TransactionDirectionOut TransactionDirection = "outgoing"

	// TransactionDirectionReconcile The transaction is an internal reconciliation transaction
	TransactionDirectionReconcile TransactionDirection = "reconcile"
)

// Transaction is an object representing the BitCoin transaction
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type Transaction struct {
	// Base model
	Model `bson:",inline"`

	// Standard transaction model base fields
	TransactionBase `bson:",inline"`

	// Model specific fields
	XpubInIDs       IDs             `json:"xpub_in_ids,omitempty" toml:"xpub_in_ids" yaml:"xpub_in_ids" gorm:"<-;type:json" bson:"xpub_in_ids,omitempty"`
	XpubOutIDs      IDs             `json:"xpub_out_ids,omitempty" toml:"xpub_out_ids" yaml:"xpub_out_ids" gorm:"<-;type:json" bson:"xpub_out_ids,omitempty"`
	BlockHash       string          `json:"block_hash" toml:"block_hash" yaml:"block_hash" gorm:"<-;type:char(64);comment:This is the related block when the transaction was mined" bson:"block_hash,omitempty"`
	BlockHeight     uint64          `json:"block_height" toml:"block_height" yaml:"block_height" gorm:"<-;type:bigint;comment:This is the related block when the transaction was mined" bson:"block_height,omitempty"`
	Fee             uint64          `json:"fee" toml:"fee" yaml:"fee" gorm:"<-create;type:bigint" bson:"fee,omitempty"`
	NumberOfInputs  uint32          `json:"number_of_inputs" toml:"number_of_inputs" yaml:"number_of_inputs" gorm:"<-;type:int" bson:"number_of_inputs,omitempty"`
	NumberOfOutputs uint32          `json:"number_of_outputs" toml:"number_of_outputs" yaml:"number_of_outputs" gorm:"<-;type:int" bson:"number_of_outputs,omitempty"`
	DraftID         string          `json:"draft_id" toml:"draft_id" yaml:"draft_id" gorm:"<-create;type:varchar(64);index;comment:This is the related draft id" bson:"draft_id,omitempty"`
	TotalValue      uint64          `json:"total_value" toml:"total_value" yaml:"total_value" gorm:"<-create;type:bigint" bson:"total_value,omitempty"`
	XpubMetadata    XpubMetadata    `json:"-" toml:"xpub_metadata" gorm:"<-;type:json;xpub_id specific metadata" bson:"xpub_metadata,omitempty"`
	XpubOutputValue XpubOutputValue `json:"-" toml:"xpub_output_value" gorm:"<-;type:json;xpub_id specific value" bson:"xpub_output_value,omitempty"`

	// Virtual Fields
	OutputValue int64                `json:"output_value" toml:"-" yaml:"-" gorm:"-" bson:"-,omitempty"`
	Status      SyncStatus           `json:"status" toml:"-" yaml:"-" gorm:"-" bson:"-"`
	Direction   TransactionDirection `json:"direction" toml:"-" yaml:"-" gorm:"-" bson:"-"`
	// Confirmations  uint64       `json:"-" toml:"-" yaml:"-" gorm:"-" bson:"-"`

	// Private for internal use
	draftTransaction   *DraftTransaction    `gorm:"-" bson:"-"` // Related draft transaction for processing and recording
	syncTransaction    *SyncTransaction     `gorm:"-" bson:"-"` // Related record if broadcast config is detected (create new recordNew)
	transactionService transactionInterface `gorm:"-" bson:"-"` // Used for interfacing methods
	utxos              []Utxo               `gorm:"-" bson:"-"` // json:"destinations,omitempty"
	xPubID             string               `gorm:"-" bson:"-"` // XPub of the user registering this transaction
	beforeCreateCalled bool                 `gorm:"-" bson:"-"` // Private information that the transaction lifecycle method BeforeCreate was already called
}

// newTransactionBase creates the standard transaction model base
func newTransactionBase(hex string, opts ...ModelOps) *Transaction {
	return &Transaction{
		TransactionBase: TransactionBase{
			Hex: hex,
		},
		Model:              *NewBaseModel(ModelTransaction, opts...),
		Status:             statusComplete,
		transactionService: transactionService{},
		XpubOutputValue:    map[string]int64{},
	}
}

// newTransaction will start a new transaction model
func newTransaction(txHex string, opts ...ModelOps) (tx *Transaction) {
	tx = newTransactionBase(txHex, opts...)

	// Set the ID
	if len(tx.Hex) > 0 {
		_ = tx.setID()
	}

	// Set xPub ID
	tx.setXPubID()

	return
}

// newTransactionWithDraftID will start a new transaction model and set the draft ID
func newTransactionWithDraftID(txHex, draftID string, opts ...ModelOps) (tx *Transaction) {
	tx = newTransaction(txHex, opts...)
	tx.DraftID = draftID
	return
}

// newTransactionFromIncomingTransaction will start a new transaction model using an incomingTx
func newTransactionFromIncomingTransaction(incomingTx *IncomingTransaction) *Transaction {

	// Create the base
	tx := newTransactionBase(incomingTx.Hex, incomingTx.GetOptions(true)...)
	tx.TransactionBase.parsedTx = incomingTx.TransactionBase.parsedTx
	tx.rawXpubKey = incomingTx.rawXpubKey
	tx.setXPubID()

	// Set the generic metadata (might be ignored if no xPub is used)
	tx.Metadata = incomingTx.Metadata

	// Set the ID (run the same method)
	if len(tx.Hex) > 0 {
		_ = tx.setID()
	}

	// Set the fields
	tx.NumberOfOutputs = uint32(len(tx.TransactionBase.parsedTx.Outputs))
	tx.NumberOfInputs = uint32(len(tx.TransactionBase.parsedTx.Inputs))
	tx.Status = statusProcessing

	return tx
}

// getTransactionByID will get the model from a given transaction ID
func getTransactionByID(ctx context.Context, xPubID, txID string, opts ...ModelOps) (*Transaction, error) {

	// Construct an empty tx
	tx := newTransaction("", opts...)
	tx.ID = txID
	tx.xPubID = xPubID

	// Get the record
	if err := Get(ctx, tx, nil, false, defaultDatabaseReadTimeout, false); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return tx, nil
}

// setXPubID will set the xPub ID on the model
func (m *Transaction) setXPubID() {
	if len(m.rawXpubKey) > 0 && len(m.xPubID) == 0 {
		m.xPubID = utils.Hash(m.rawXpubKey)
	}
}

// getTransactions will get all the transactions with the given conditions
func getTransactions(ctx context.Context, metadata *Metadata, conditions *map[string]interface{},
	queryParams *datastore.QueryParams, opts ...ModelOps) ([]*Transaction, error) {

	modelItems := make([]*Transaction, 0)
	if err := getModelsByConditions(ctx, ModelTransaction, &modelItems, metadata, conditions, queryParams, opts...); err != nil {
		return nil, err
	}

	return modelItems, nil
}

// getTransactionsAggregate will get a count of all transactions per aggregate column with the given conditions
func getTransactionsAggregate(ctx context.Context, metadata *Metadata, conditions *map[string]interface{},
	aggregateColumn string, opts ...ModelOps) (map[string]interface{}, error) {

	modelItems := make([]*Transaction, 0)
	results, err := getModelsAggregateByConditions(
		ctx, ModelTransaction, &modelItems, metadata, conditions, aggregateColumn, opts...,
	)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// getTransactionsCount will get a count of all the transactions with the given conditions
func getTransactionsCount(ctx context.Context, metadata *Metadata, conditions *map[string]interface{},
	opts ...ModelOps) (int64, error) {
	return getModelCountByConditions(ctx, ModelTransaction, Transaction{}, metadata, conditions, opts...)
}

// getTransactionsCountByXpubID will get the count of all the models for a given xpub ID
func getTransactionsCountByXpubID(ctx context.Context, xPubID string, metadata *Metadata,
	conditions *map[string]interface{}, opts ...ModelOps) (int64, error) {

	dbConditions := processDBConditions(xPubID, conditions, metadata)

	return getTransactionsCountInternal(ctx, dbConditions, opts...)
}

// getTransactionsByXpubID will get all the models for a given xpub ID
func getTransactionsByXpubID(ctx context.Context, xPubID string,
	metadata *Metadata, conditions *map[string]interface{},
	queryParams *datastore.QueryParams, opts ...ModelOps) ([]*Transaction, error) {

	dbConditions := processDBConditions(xPubID, conditions, metadata)

	return getTransactionsInternal(ctx, dbConditions, xPubID, queryParams, opts...)
}

func processDBConditions(xPubID string, conditions *map[string]interface{},
	metadata *Metadata) map[string]interface{} {

	dbConditions := map[string]interface{}{
		"$or": []map[string]interface{}{{
			"xpub_in_ids": xPubID,
		}, {
			"xpub_out_ids": xPubID,
		}},
	}

	// check for direction query
	if conditions != nil && (*conditions)["direction"] != nil {
		direction := (*conditions)["direction"].(string)
		if direction == string(TransactionDirectionIn) {
			dbConditions["xpub_output_value"] = map[string]interface{}{
				xPubID: map[string]interface{}{
					"$gt": 0,
				},
			}
		} else if direction == string(TransactionDirectionOut) {
			dbConditions["xpub_output_value"] = map[string]interface{}{
				xPubID: map[string]interface{}{
					"$lt": 0,
				},
			}
		} else if direction == string(TransactionDirectionReconcile) {
			dbConditions["xpub_output_value"] = map[string]interface{}{
				xPubID: 0,
			}
		}
		delete(*conditions, "direction")
	}

	if metadata != nil && len(*metadata) > 0 {
		and := make([]map[string]interface{}, 0)
		if _, ok := dbConditions["$and"]; ok {
			and = dbConditions["$and"].([]map[string]interface{})
		}
		for key, value := range *metadata {
			condition := map[string]interface{}{
				"$or": []map[string]interface{}{{
					metadataField: Metadata{
						key: value,
					},
				}, {
					xPubMetadataField: map[string]interface{}{
						xPubID: Metadata{
							key: value,
						},
					},
				}},
			}
			and = append(and, condition)
		}
		dbConditions["$and"] = and
	}

	if conditions != nil && len(*conditions) > 0 {
		and := make([]map[string]interface{}, 0)
		if _, ok := dbConditions["$and"]; ok {
			and = dbConditions["$and"].([]map[string]interface{})
		}
		and = append(and, *conditions)
		dbConditions["$and"] = and
	}

	return dbConditions
}

// getTransactionsInternal get all transactions for the given conditions
// NOTE: this function should only be used internally, it allows to query the whole transaction table
func getTransactionsInternal(ctx context.Context, conditions map[string]interface{}, xPubID string,
	queryParams *datastore.QueryParams, opts ...ModelOps) ([]*Transaction, error) {
	var models []Transaction
	if err := getModels(
		ctx,
		NewBaseModel(ModelTransaction, opts...).Client().Datastore(),
		&models,
		conditions,
		queryParams,
		defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	// Loop and enrich
	transactions := make([]*Transaction, 0)
	for index := range models {
		models[index].enrich(ModelTransaction, opts...)
		models[index].xPubID = xPubID
		tx := &models[index]
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// getTransactionsCountInternal get a count of all transactions for the given conditions
func getTransactionsCountInternal(ctx context.Context, conditions map[string]interface{},
	opts ...ModelOps) (int64, error) {

	count, err := getModelCount(
		ctx,
		NewBaseModel(ModelNameEmpty, opts...).Client().Datastore(),
		Transaction{},
		conditions,
		defaultDatabaseReadTimeout,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateTransactionMetadata will update the transaction metadata by xPubID
func (m *Transaction) UpdateTransactionMetadata(xPubID string, metadata Metadata) error {
	if xPubID == "" {
		return ErrXpubIDMisMatch
	}

	// transaction metadata is saved per xPubID
	if m.XpubMetadata == nil {
		m.XpubMetadata = make(XpubMetadata)
	}
	if m.XpubMetadata[xPubID] == nil {
		m.XpubMetadata[xPubID] = make(Metadata)
	}

	for key, value := range metadata {
		if value == nil {
			delete(m.XpubMetadata[xPubID], key)
		} else {
			m.XpubMetadata[xPubID][key] = value
		}
	}

	return nil
}

// GetModelName will get the name of the current model
func (m *Transaction) GetModelName() string {
	return ModelTransaction.String()
}

// GetModelTableName will get the db table name of the current model
func (m *Transaction) GetModelTableName() string {
	return tableTransactions
}

// Save will save the model into the Datastore
func (m *Transaction) Save(ctx context.Context) (err error) {

	// Prepare the metadata
	if len(m.Metadata) > 0 {
		// set the metadata to be xpub specific, but only if we have a valid xpub ID
		if m.xPubID != "" {
			// was metadata set via opts ?
			if m.XpubMetadata == nil {
				m.XpubMetadata = make(XpubMetadata)
			}
			if _, ok := m.XpubMetadata[m.xPubID]; !ok {
				m.XpubMetadata[m.xPubID] = make(Metadata)
			}
			for key, value := range m.Metadata {
				m.XpubMetadata[m.xPubID][key] = value
			}
			// todo will this overwrite the global metadata ?
			m.Metadata = nil
		} else {
			m.DebugLog("xPub id is missing from transaction, cannot store metadata")
		}
	}

	return Save(ctx, m)
}

// GetID will get the ID
func (m *Transaction) GetID() string {
	return m.ID
}

// setID will set the ID from the transaction hex
func (m *Transaction) setID() (err error) {

	// Parse the hex (if not already parsed)
	if m.TransactionBase.parsedTx == nil {
		if m.TransactionBase.parsedTx, err = bt.NewTxFromString(m.Hex); err != nil {
			return
		}
	}

	// Set the true transaction ID
	m.ID = m.TransactionBase.parsedTx.TxID()

	return
}

// getValue calculates the value of the transaction
func (m *Transaction) getValues() (outputValue uint64, fee uint64) {

	// Parse the outputs
	for _, output := range m.TransactionBase.parsedTx.Outputs {
		outputValue += output.Satoshis
	}

	// Remove the "change" from the transaction if found
	// todo: this will NOT work for an "external" tx that is coming into our system?
	if m.draftTransaction != nil {
		outputValue -= m.draftTransaction.Configuration.ChangeSatoshis
		fee = m.draftTransaction.Configuration.Fee
	} else { // external transaction

		// todo: missing inputs in some tests?
		var inputValue uint64
		for _, input := range m.TransactionBase.parsedTx.Inputs {
			inputValue += input.PreviousTxSatoshis
		}

		if inputValue > 0 {
			fee = inputValue - outputValue
			outputValue -= fee
		}

		// todo: outputs we know are accumulated
	}

	// remove the fee from the value
	if outputValue > fee {
		outputValue -= fee
	}

	return
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *Transaction) BeforeCreating(ctx context.Context) error {

	if m.beforeCreateCalled {
		m.DebugLog("skipping: " + m.Name() + " BeforeCreating hook, because already called")
		return nil
	}

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Test for required field(s)
	if len(m.Hex) == 0 {
		return ErrMissingFieldHex
	}

	// Set the xPubID
	m.setXPubID()

	// Set the ID - will also parse and verify the tx
	err := m.setID()
	if err != nil {
		return err
	}

	// 	m.xPubID is the xpub of the user registering the transaction
	if len(m.xPubID) > 0 && len(m.DraftID) > 0 {

		// Only get the draft if we haven't already
		if m.draftTransaction == nil {
			if m.draftTransaction, err = getDraftTransactionID(
				ctx, m.xPubID, m.DraftID, m.GetOptions(false)...,
			); err != nil {
				return err
			} else if m.draftTransaction == nil {
				return ErrDraftNotFound
			}
		}
	}

	// Validations and broadcast config check
	if m.draftTransaction != nil {

		// No config set? Use the default from the client
		if m.draftTransaction.Configuration.Sync == nil {
			m.draftTransaction.Configuration.Sync = m.Client().DefaultSyncConfig()
		}

		// Create the sync transaction model
		sync := newSyncTransaction(
			m.GetID(),
			m.draftTransaction.Configuration.Sync,
			m.GetOptions(true)...,
		)

		// Found any p2p outputs?
		p2pStatus := SyncStatusSkipped
		if m.draftTransaction.Configuration.Outputs != nil {
			for _, output := range m.draftTransaction.Configuration.Outputs {
				if output.PaymailP4 != nil && output.PaymailP4.ResolutionType == ResolutionTypeP2P {
					p2pStatus = SyncStatusPending
				}
			}
		}
		sync.P2PStatus = p2pStatus

		// Use the same metadata
		sync.Metadata = m.Metadata

		// set this transaction on the sync transaction object. This is needed for the first broadcast
		sync.transaction = m

		// If all the options are skipped, do not make a new model (ignore the record)
		if !sync.isSkipped() {
			m.syncTransaction = sync
		}
	}

	// If we are external and the user disabled incoming transaction checking, check outputs
	if m.isExternal() && !m.Client().IsITCEnabled() {
		// Check that the transaction has >= 1 known destination
		if !m.TransactionBase.hasOneKnownDestination(ctx, m.Client(), m.GetOptions(false)...) {
			return ErrNoMatchingOutputs
		}
	}

	// Process the UTXOs
	if err = m.processUtxos(ctx); err != nil {
		return err
	}

	// Set the values from the inputs/outputs and draft tx
	m.TotalValue, m.Fee = m.getValues()

	// Add values if found
	if m.TransactionBase.parsedTx != nil {
		m.NumberOfInputs = uint32(len(m.TransactionBase.parsedTx.Inputs))
		m.NumberOfOutputs = uint32(len(m.TransactionBase.parsedTx.Outputs))
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	m.beforeCreateCalled = true
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *Transaction) AfterCreated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	// Pre-build the options
	opts := m.GetOptions(false)

	// update the xpub balances
	for xPubID, balance := range m.XpubOutputValue {
		// todo: run this in a go routine? (move this into a function on the xpub model?)
		xPub, err := getXpubWithCache(ctx, m.Client(), "", xPubID, opts...)
		if err != nil {
			return err
		} else if xPub == nil {
			return ErrMissingRequiredXpub
		}
		if err = xPub.incrementBalance(ctx, balance); err != nil {
			return err
		}
	}

	// Update the draft transaction, process broadcasting
	// todo: go routine (however it's not working, panic in save for missing datastore)
	if m.draftTransaction != nil {
		m.draftTransaction.Status = DraftStatusComplete
		m.draftTransaction.FinalTxID = m.ID
		if err := m.draftTransaction.Save(ctx); err != nil {
			return err
		}
	}

	// Fire notifications (this is already in a go routine)
	notify(notifications.EventTypeCreate, m)

	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

// AfterUpdated will fire after the model is updated in the Datastore
func (m *Transaction) AfterUpdated(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterUpdated hook...")

	// Fire notifications (this is already in a go routine)
	notify(notifications.EventTypeUpdate, m)

	m.DebugLog("end: " + m.Name() + " AfterUpdated hook")
	return nil
}

// AfterDeleted will fire after the model is deleted in the Datastore
func (m *Transaction) AfterDeleted(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterDelete hook...")

	// Fire notifications (this is already in a go routine)
	notify(notifications.EventTypeDelete, m)

	m.DebugLog("end: " + m.Name() + " AfterDelete hook")
	return nil
}

// ChildModels will get any related sub models
func (m *Transaction) ChildModels() (childModels []ModelInterface) {

	// Add the UTXOs if found
	for index := range m.utxos {
		childModels = append(childModels, &m.utxos[index])
	}

	// Add the broadcast transaction record
	if m.syncTransaction != nil {
		childModels = append(childModels, m.syncTransaction)
	}

	return
}

// processUtxos will process the inputs and outputs for UTXOs
func (m *Transaction) processUtxos(ctx context.Context) error {

	// Input should be processed only for outcomming transactions
	if m.draftTransaction != nil {
		if err := m.processInputs(ctx); err != nil {
			return err
		}
	}

	return m.processOutputs(ctx)
}

// processTxOutputs will process the transaction outputs
func (m *Transaction) processOutputs(ctx context.Context) (err error) {

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

func (m *Transaction) isExternal() bool {
	return m.draftTransaction == nil
}

// processTxInputs will process the transaction inputs
func (m *Transaction) processInputs(ctx context.Context) (err error) {

	// Pre-build the options
	opts := m.GetOptions(false)
	client := m.Client()

	var utxo *Utxo

	// check whether we are spending an internal utxo
	for index := range m.TransactionBase.parsedTx.Inputs {

		// todo: optimize this SQL SELECT to get all utxos in one query?
		if utxo, err = m.transactionService.getUtxo(ctx,
			hex.EncodeToString(m.TransactionBase.parsedTx.Inputs[index].PreviousTxID()),
			m.TransactionBase.parsedTx.Inputs[index].PreviousTxOutIndex,
			opts...,
		); err != nil {
			return
		} else if utxo != nil { // Found a UTXO record

			// Is Spent?
			if len(utxo.SpendingTxID.String) > 0 {
				return ErrUtxoAlreadySpent
			}

			// Only if IUC is enabled (or client is nil which means its enabled by default)
			if client == nil || client.IsIUCEnabled() {

				// check whether the utxo is spent
				isReserved := len(utxo.DraftID.String) > 0
				matchesDraft := m.draftTransaction != nil && utxo.DraftID.String == m.draftTransaction.ID

				// Check whether the spending transaction was reserved by the draft transaction (in the utxo)
				if !isReserved {
					return ErrUtxoNotReserved
				}
				if !matchesDraft {
					return ErrDraftIDMismatch
				}
			}

			// Update the output value
			if _, ok := m.XpubOutputValue[utxo.XpubID]; !ok {
				m.XpubOutputValue[utxo.XpubID] = 0
			}
			m.XpubOutputValue[utxo.XpubID] -= int64(utxo.Satoshis)

			// Mark utxo as spent
			utxo.SpendingTxID.Valid = true
			utxo.SpendingTxID.String = m.ID
			m.utxos = append(m.utxos, *utxo)

			// Add the xPub ID
			if !utils.StringInSlice(utxo.XpubID, m.XpubInIDs) {
				m.XpubInIDs = append(m.XpubInIDs, utxo.XpubID)
			}
		}

		// todo: what if the utxo is nil (not found)?
	}

	return
}

// IsXpubAssociated will check if this key is associated to this transaction
func (m *Transaction) IsXpubAssociated(rawXpubKey string) bool {

	// Hash the raw key
	xPubID := utils.Hash(rawXpubKey)
	return m.IsXpubIDAssociated(xPubID)
}

// IsXpubIDAssociated will check if an xPub ID is associated
func (m *Transaction) IsXpubIDAssociated(xPubID string) bool {
	if len(xPubID) == 0 {
		return false
	}

	// On the input side
	for _, id := range m.XpubInIDs {
		if id == xPubID {
			return true
		}
	}

	// On the output side
	for _, id := range m.XpubOutIDs {
		if id == xPubID {
			return true
		}
	}
	return false
}

// Display filter the model for display
func (m *Transaction) Display() interface{} {

	// In case it was not set
	m.setXPubID()

	if len(m.XpubMetadata) > 0 && len(m.XpubMetadata[m.xPubID]) > 0 {
		if m.Metadata == nil {
			m.Metadata = make(Metadata)
		}
		for key, value := range m.XpubMetadata[m.xPubID] {
			m.Metadata[key] = value
		}
	}

	m.OutputValue = int64(0)
	if len(m.XpubOutputValue) > 0 && m.XpubOutputValue[m.xPubID] != 0 {
		m.OutputValue = m.XpubOutputValue[m.xPubID]
	}

	if m.OutputValue > 0 {
		m.Direction = TransactionDirectionIn
	} else {
		m.Direction = TransactionDirectionOut
	}

	m.XpubInIDs = nil
	m.XpubOutIDs = nil
	m.XpubMetadata = nil
	m.XpubOutputValue = nil
	return m
}

// Migrate model specific migration on startup
func (m *Transaction) Migrate(client datastore.ClientInterface) error {

	tableName := client.GetTableName(tableTransactions)
	if client.Engine() == datastore.MySQL {
		if err := m.migrateMySQL(client, tableName); err != nil {
			return err
		}
	} else if client.Engine() == datastore.PostgreSQL {
		if err := m.migratePostgreSQL(client, tableName); err != nil {
			return err
		}
	}

	return client.IndexMetadata(tableName, xPubMetadataField)
}

// migratePostgreSQL is specific migration SQL for Postgresql
func (m *Transaction) migratePostgreSQL(client datastore.ClientInterface, tableName string) error {

	tx := client.Execute(`CREATE INDEX IF NOT EXISTS idx_` + tableName + `_xpub_in_ids ON ` +
		tableName + ` USING gin (xpub_in_ids jsonb_ops)`)
	if tx.Error != nil {
		return tx.Error
	}

	if tx = client.Execute(`CREATE INDEX IF NOT EXISTS idx_` + tableName + `_xpub_out_ids ON ` +
		tableName + ` USING gin (xpub_out_ids jsonb_ops)`); tx.Error != nil {
		return tx.Error
	}

	return nil
}

// migrateMySQL is specific migration SQL for MySQL
func (m *Transaction) migrateMySQL(client datastore.ClientInterface, tableName string) error {

	idxName := "idx_" + tableName + "_xpub_in_ids"
	idxExists, err := client.IndexExists(tableName, idxName)
	if err != nil {
		return err
	}
	if !idxExists {
		tx := client.Execute("ALTER TABLE `" + tableName + "`" +
			" ADD INDEX " + idxName + " ( (CAST(xpub_in_ids AS CHAR(64) ARRAY)) )")
		if tx.Error != nil {
			m.Client().Logger().Error(context.Background(), "failed creating json index on mysql: "+tx.Error.Error())
			return nil //nolint:nolintlint,nilerr // error is not needed
		}
	}

	idxName = "idx_" + tableName + "_xpub_out_ids"
	if idxExists, err = client.IndexExists(
		tableName, idxName,
	); err != nil {
		return err
	}
	if !idxExists {
		tx := client.Execute("ALTER TABLE `" + tableName + "`" +
			" ADD INDEX " + idxName + " ( (CAST(xpub_out_ids AS CHAR(64) ARRAY)) )")
		if tx.Error != nil {
			m.Client().Logger().Error(context.Background(), "failed creating json index on mysql: "+tx.Error.Error())
			return nil //nolint:nolintlint,nilerr // error is not needed
		}
	}

	tx := client.Execute("ALTER TABLE `" + tableName + "` MODIFY COLUMN hex longtext")
	if tx.Error != nil {
		m.Client().Logger().Error(context.Background(), "failed changing hex type to longtext in MySQL: "+tx.Error.Error())
		return nil //nolint:nolintlint,nilerr // error is not needed
	}

	return nil
}

// hasOneKnownDestination will check if the transaction has at least one known destination
//
// This is used to validate if an external transaction should be recorded into the engine
func (m *TransactionBase) hasOneKnownDestination(ctx context.Context, client ClientInterface, opts ...ModelOps) bool {

	// todo: this can be optimized searching X records at a time vs loop->query->loop->query
	lockingScript := ""
	for index := range m.parsedTx.Outputs {
		lockingScript = m.parsedTx.Outputs[index].LockingScript.String()
		destination, err := getDestinationWithCache(ctx, client, "", "", lockingScript, opts...)
		if err != nil {
			destination = newDestination("", lockingScript, opts...)
			destination.Client().Logger().Error(ctx, "error getting destination: "+err.Error())
		} else if destination != nil && destination.LockingScript == lockingScript {
			return true
		}
	}
	return false
}
