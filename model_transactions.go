package bux

import (
	"context"

	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
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
	MerkleProof     MerkleProof     `json:"merkle_proof" toml:"merkle_proof" yaml:"merkle_proof" gorm:"<-;type:text;comment:Merkle Proof payload from mAPI" bson:"merkle_proof,omitempty"`
	BUMP            BUMP            `json:"bump" toml:"bump" yaml:"bump" gorm:"<-;type:text;comment:BSV Unified Merkle Path (BUMP) Format" bson:"bump,omitempty"`

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
	XPubID             string               `gorm:"-" bson:"-"` // XPub of the user registering this transaction
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

	return tx
}

// setXPubID will set the xPub ID on the model
func (m *Transaction) setXPubID() {
	if len(m.rawXpubKey) > 0 && len(m.XPubID) == 0 {
		m.XPubID = utils.Hash(m.rawXpubKey)
	}
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

func (m *Transaction) isExternal() bool {
	return m.draftTransaction == nil
}

func (m *Transaction) updateChainInfo(txInfo *chainstate.TransactionInfo) {
	m.BlockHash = txInfo.BlockHash
	m.BlockHeight = uint64(txInfo.BlockHeight)

	if txInfo.MerkleProof != nil {
		mp := MerkleProof(*txInfo.MerkleProof)
		m.MerkleProof = mp

		bump := mp.ToBUMP()
		bump.BlockHeight = m.BlockHeight
		m.BUMP = bump
	}
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

	if len(m.XpubMetadata) > 0 && len(m.XpubMetadata[m.XPubID]) > 0 {
		if m.Metadata == nil {
			m.Metadata = make(Metadata)
		}
		for key, value := range m.XpubMetadata[m.XPubID] {
			m.Metadata[key] = value
		}
	}

	m.OutputValue = int64(0)
	if len(m.XpubOutputValue) > 0 && m.XpubOutputValue[m.XPubID] != 0 {
		m.OutputValue = m.XpubOutputValue[m.XPubID]
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

// RegisterTasks will register the model specific tasks on client initialization
func (m *Transaction) RegisterTasks() error {
	// No task manager loaded?
	tm := m.Client().Taskmanager()
	if tm == nil {
		return nil
	}

	ctx := context.Background()
	checkTask := m.Name() + "_" + TransactionActionCheck

	if err := tm.RegisterTask(&taskmanager.Task{
		Name:       checkTask,
		RetryLimit: 1,
		Handler: func(client ClientInterface) error {
			if taskErr := taskCheckTransactions(ctx, client.Logger(), WithClient(client)); taskErr != nil {
				client.Logger().Error(ctx, "error running "+checkTask+" task: "+taskErr.Error())
			}
			return nil
		},
	}); err != nil {
		return err
	}

	return tm.RunTask(ctx, &taskmanager.TaskOptions{
		Arguments:      []interface{}{m.Client()},
		RunEveryPeriod: m.Client().GetTaskPeriod(checkTask),
		TaskName:       checkTask,
	})
}
