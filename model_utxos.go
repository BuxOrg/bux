package bux

import (
	"context"
	"fmt"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
	"github.com/pkg/errors"
)

// UtxoPointer is the actual pointer (index) for the UTXO
type UtxoPointer struct {
	TransactionID string `json:"transaction_id" toml:"transaction_id" yaml:"transaction_id" gorm:"<-:create;type:char(64);index;comment:This is the id of the related transaction" bson:"transaction_id"`
	OutputIndex   uint32 `json:"output_index" toml:"output_index" yaml:"output_index" gorm:"<-:create;type:uint;comment:This is the index of the output in the transaction" bson:"output_index"`
}

// Utxo is an object representing a BitCoin unspent transaction
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type Utxo struct {
	// Base model
	Model `bson:",inline"`

	// Standard utxo model base fields
	UtxoPointer `bson:",inline"`

	// Model specific fields
	ID           string           `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the sha256 hash of the (<txid>|vout)" bson:"_id"`
	XpubID       string           `json:"xpub_id" toml:"xpub_id" yaml:"xpub_id" gorm:"<-:create;type:char(64);index;comment:This is the related xPub" bson:"xpub_id"`
	Satoshis     uint64           `json:"satoshis" toml:"satoshis" yaml:"satoshis" gorm:"<-:create;type:uint;comment:This is the amount of satoshis in the output" bson:"satoshis"`
	ScriptPubKey string           `json:"script_pub_key" toml:"script_pub_key" yaml:"script_pub_key" gorm:"<-:create;type:text;comment:This is the script pub key" bson:"script_pub_key"`
	Type         string           `json:"type" toml:"type" yaml:"type" gorm:"<-:create;type:text;comment:Type of output" bson:"type"`
	DraftID      utils.NullString `json:"draft_id" toml:"draft_id" yaml:"draft_id" gorm:"<-;type:varchar(64);index;comment:Related draft id for reservations" bson:"draft_id,omitempty"`
	ReservedAt   utils.NullTime   `json:"reserved_at" toml:"reserved_at" yaml:"reserved_at" gorm:"<-;comment:When it was reserved" bson:"reserved_at,omitempty"`
	SpendingTxID utils.NullString `json:"spending_tx_id,omitempty" toml:"spending_tx_id" yaml:"spending_tx_id" gorm:"<-;type:char(64);index;comment:This is tx ID of the spend" bson:"spending_tx_id,omitempty"`
}

// newUtxo will start a new utxo model
func newUtxo(xPubID, txID, scriptPubKey string, index uint32, satoshis uint64, opts ...ModelOps) *Utxo {
	return &Utxo{
		UtxoPointer: UtxoPointer{
			OutputIndex:   index,
			TransactionID: txID,
		},
		Model:        *NewBaseModel(ModelUtxo, opts...),
		Satoshis:     satoshis,
		ScriptPubKey: scriptPubKey,
		XpubID:       xPubID,
	}
}

// GetSpendableUtxos Get all spendable utxos - yes really!
func GetSpendableUtxos(ctx context.Context, xPubID, utxoType string, fromUtxos []*UtxoPointer, opts ...ModelOps) ([]*Utxo, error) {

	// Construct the conditions and results
	var models []Utxo
	conditions := map[string]interface{}{
		xPubIDField:       xPubID,
		typeField:         utxoType,
		draftIDField:      nil,
		spendingTxIDField: nil,
	}

	if fromUtxos != nil {
		for _, fromUtxo := range fromUtxos {
			utxo, err := getUtxo(ctx, fromUtxo.TransactionID, fromUtxo.OutputIndex, opts...)
			if err != nil {
				return nil, err
			}
			if utxo.XpubID != xPubID || utxo.SpendingTxID.Valid {
				return nil, ErrUtxoAlreadySpent
			}
			models = append(models, *utxo)
		}
	} else {
		// Get the records
		if err := getModels(
			ctx, NewBaseModel(ModelNameEmpty, opts...).Client().Datastore(),
			&models, conditions, 0, 0, "", "", defaultDatabaseReadTimeout,
		); err != nil {
			if errors.Is(err, datastore.ErrNoResults) {
				return nil, nil
			}
			return nil, err
		}
	}

	// Loop and enrich
	utxos := make([]*Utxo, 0)
	for index := range models {
		models[index].enrich(ModelUtxo, opts...)
		utxos = append(utxos, &models[index])
	}

	return utxos, nil
}

// UnReserveUtxos remove the reservation on the utxos for the given draft ID
func UnReserveUtxos(ctx context.Context, xPubID, draftID string, opts ...ModelOps) error {
	var models []Utxo
	conditions := map[string]interface{}{
		xPubIDField:  xPubID,
		draftIDField: draftID,
	}

	// Get the records
	if err := getModels(
		ctx, NewBaseModel(ModelNameEmpty, opts...).Client().Datastore(),
		&models, conditions, 0, 0, "", "", defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil
		}
		return err
	}

	// Loop and un-reserve
	for index := range models {
		utxo := models[index]
		utxo.enrich(ModelUtxo, opts...)
		utxo.DraftID.Valid = false
		utxo.ReservedAt.Valid = false
		if err := utxo.Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

// ReserveUtxos reserve utxos for the given draft ID and amount
func ReserveUtxos(ctx context.Context, xPubID, draftID string,
	satoshis uint64, feePerByte float64, fromUtxos []*UtxoPointer, opts ...ModelOps) ([]*Utxo, error) {

	// Create base model
	m := NewBaseModel(ModelNameEmpty, opts...)

	// Create the lock and set the release for after the function completes
	unlock, err := newWaitWriteLock(
		ctx, fmt.Sprintf(lockKeyReserveUtxo, xPubID), m.Client().Cachestore(),
	)
	defer unlock()
	if err != nil {
		return nil, err
	}

	// Get spendable utxos
	// todo improve this to not Get all utxos - make smarter
	var freeUtxos []*Utxo
	if freeUtxos, err = GetSpendableUtxos(
		ctx, xPubID, utils.ScriptTypePubKeyHash, fromUtxos, opts..., // todo: allow reservation of utxos by a different utxo destination type
	); err != nil {
		return nil, err
	}

	// Set vars
	size := utils.GetInputSizeForType(utils.ScriptTypePubKeyHash)
	feeNeeded := uint64(0)
	reservedSatoshis := uint64(0)
	utxos := new([]*Utxo)

	// Loop the returned utxos
	for _, utxo := range freeUtxos {

		// Set the values on the UTXO
		utxo.DraftID.Valid = true
		utxo.DraftID.String = draftID
		utxo.ReservedAt.Valid = true
		utxo.ReservedAt.Time = time.Now().UTC()

		// Accumulate the reserved satoshis
		reservedSatoshis += utxo.Satoshis

		// Save the UTXO
		// todo: should occur in 1 DB transaction
		if err = utxo.Save(ctx); err != nil {
			return nil, err
		}

		// Add the utxo to the final slice
		*utxos = append(*utxos, utxo)

		// add fee for this new input
		feeNeeded += uint64(float64(size) * feePerByte)
		if reservedSatoshis >= (satoshis + feeNeeded) {
			break
		}
	}

	if reservedSatoshis < satoshis {
		if err = UnReserveUtxos(
			ctx, xPubID, draftID, m.GetOptions(false)...,
		); err != nil {
			return nil, errors.Wrap(err, ErrNotEnoughUtxos.Error())
		}
		return nil, ErrNotEnoughUtxos
	}

	// todo: return error if no UTXOs found or saved?

	return *utxos, nil
}

// newUtxoFromTxID will start a new utxo model
func newUtxoFromTxID(txID string, index uint32, opts ...ModelOps) *Utxo {
	return &Utxo{
		Model: *NewBaseModel(ModelUtxo, opts...),
		UtxoPointer: UtxoPointer{
			OutputIndex:   index,
			TransactionID: txID,
		},
	}
}

// getUtxosByXpubID
func getUtxosByXpubID(ctx context.Context, xpubID string, pageSize, page int, orderByField,
	sortDirection string, opts ...ModelOps) ([]*Utxo, error) {
	conditions := map[string]interface{}{
		xPubIDField: xpubID,
	}
	return getUtxosByConditions(ctx, conditions, pageSize, page, orderByField, sortDirection, opts...)
}

// getUtxosByDraftID
func getUtxosByDraftID(ctx context.Context, draftID string, pageSize, page int,
	orderByField, sortDirection string, opts ...ModelOps) ([]*Utxo, error) {
	conditions := map[string]interface{}{
		draftIDField: draftID,
	}
	return getUtxosByConditions(ctx, conditions, pageSize, page, orderByField, sortDirection, opts...)
}

func getUtxosByConditions(ctx context.Context, conditions map[string]interface{}, pageSize,
	page int, orderByField, sortDirection string, opts ...ModelOps) ([]*Utxo, error) {
	var models []Utxo
	if err := getModels(
		ctx, NewBaseModel(
			ModelNameEmpty, opts...).Client().Datastore(),
		&models, conditions, pageSize, page, orderByField, sortDirection, databaseLongReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	// Loop and enrich
	utxos := make([]*Utxo, 0)
	for index := range models {
		models[index].enrich(ModelUtxo, opts...)
		utxos = append(utxos, &models[index])
	}
	return utxos, nil
}

// getUtxo will get the utxo with the given conditions
func getUtxo(ctx context.Context, txID string, index uint32, opts ...ModelOps) (*Utxo, error) {

	// Start the new model
	utxo := newUtxoFromTxID(txID, index, opts...)

	// Create the conditions for searching
	conditions := map[string]interface{}{
		"transaction_id": txID,
		"output_index":   index,
	}

	// Get the records
	if err := Get(ctx, utxo, conditions, true, defaultDatabaseReadTimeout); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return utxo, nil
}

// GetModelName will get the name of the current model
func (m *Utxo) GetModelName() string {
	return ModelUtxo.String()
}

// GetModelTableName will get the db table name of the current model
func (m *Utxo) GetModelTableName() string {
	return tableUTXOs
}

// Save will Save the model into the Datastore
func (m *Utxo) Save(ctx context.Context) (err error) {
	return Save(ctx, m)
}

// GetID will get the ID
func (m *Utxo) GetID() string {
	if m.ID == "" {
		m.ID = m.GenerateID()
	}
	return m.ID
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *Utxo) BeforeCreating(_ context.Context) error {

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Test for required field(s)
	if len(m.ScriptPubKey) == 0 {
		return ErrMissingFieldScriptPubKey
	} else if m.Satoshis == 0 {
		return ErrMissingFieldSatoshis
	} else if len(m.TransactionID) == 0 {
		return ErrMissingFieldTransactionID
	}

	if len(m.XpubID) == 0 {
		return ErrMissingFieldXpubID
	}

	// Set the new pointer?
	/*
		if m.parsedUtxo == nil {
			m.parsedUtxo = New(bt.UTXO)
		}

		// Parse the UTXO (tx id)
		if m.parsedUtxo.TxID, err = hex.DecodeString(
			m.TransactionID,
		); err != nil {
			return err
		}

		// Parse the UTXO (locking script)
		if m.parsedUtxo.LockingScript, err = bscript2.NewFromHexString(
			m.ScriptPubKey,
		); err != nil {
			return err
		}
		m.parsedUtxo.Satoshis = m.Satoshis
		m.parsedUtxo.Vout = m.OutputIndex
	*/

	// Set the ID
	m.ID = m.GenerateID()
	m.Type = utils.GetDestinationType(m.ScriptPubKey)

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// GenerateID will generate the id of the UTXO record based on the format: <txid>|<output_index>
func (m *Utxo) GenerateID() string {
	return utils.Hash(fmt.Sprintf("%s|%d", m.TransactionID, m.OutputIndex))
}

// Migrate model specific migration on startup
func (m *Utxo) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableUTXOs), metadataField)
}
