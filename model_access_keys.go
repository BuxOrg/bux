package bux

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
)

// AccessKey is an object representing the access key
//
// An AccessKey is a private key with a corresponding public key
// The public key is hashed and saved in this model for retrieval.
// When a request is made with an access key, the public key is sent in the headers, together with
// a signature (like normally done with xPriv signing)
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type AccessKey struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	ID        string         `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique access key id" bson:"_id"`
	XpubID    string         `json:"xpub_id" toml:"xpub_id" yaml:"hash" gorm:"<-:create;type:char(64);index;comment:This is the related xPub id" bson:"xpub_id"`
	RevokedAt utils.NullTime `json:"revoked_at" toml:"revoked_at" yaml:"revoked_at" gorm:"<-;comment:When the key was revoked" bson:"revoked_at,omitempty"`

	// Private fields
	Key string `json:"key" gorm:"-" bson:"-"` // Used on "CREATE", shown to the user "once" only
}

// newAccessKey will start a new model
func newAccessKey(xPubID string, opts ...ModelOps) *AccessKey {

	privateKey, _ := bitcoin.CreatePrivateKey()
	publicKey := hex.EncodeToString(privateKey.PubKey().SerialiseCompressed())
	id := utils.Hash(publicKey)

	return &AccessKey{
		ID:     id,
		Model:  *NewBaseModel(ModelAccessKey, opts...),
		XpubID: xPubID,
		RevokedAt: utils.NullTime{NullTime: sql.NullTime{
			Valid: false,
		}},
		Key: hex.EncodeToString(privateKey.Serialise()),
	}
}

// GetAccessKey will get the model with a given ID
func GetAccessKey(ctx context.Context, id string, opts ...ModelOps) (*AccessKey, error) {

	// Construct an empty tx
	key := &AccessKey{
		ID: id,
	}
	key.enrich(ModelAccessKey, opts...)

	// Get the record
	if err := Get(ctx, key, nil, false, defaultDatabaseReadTimeout); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}
	return key, nil
}

// GetAccessKeys will get all the access keys that match the metadata search
func GetAccessKeys(ctx context.Context, xPubID string, metadata *Metadata, opts ...ModelOps) ([]*AccessKey, error) {

	// Construct an empty model
	var models []AccessKey
	conditions := map[string]interface{}{
		xPubIDField: xPubID,
	}

	if metadata != nil {
		conditions[metadataField] = metadata
	}

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

	// Loop and enrich
	accessKeys := make([]*AccessKey, 0)
	for index := range models {
		models[index].enrich(ModelDestination, opts...)
		accessKeys = append(accessKeys, &models[index])
	}

	return accessKeys, nil
}

// GetModelName will get the name of the current model
func (m *AccessKey) GetModelName() string {
	return ModelAccessKey.String()
}

// GetModelTableName will get the db table name of the current model
func (m *AccessKey) GetModelTableName() string {
	return tableAccessKeys
}

// Save will Save the model into the Datastore
func (m *AccessKey) Save(ctx context.Context) error {
	return Save(ctx, m)
}

// GetID will get the ID
func (m *AccessKey) GetID() string {
	return m.ID
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *AccessKey) BeforeCreating(_ context.Context) error {
	m.DebugLog("starting: [" + m.name.String() + "] BeforeCreating hook...")

	// Make sure ID is valid
	if len(m.ID) == 0 {
		return ErrMissingFieldID
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// RegisterTasks will register the model specific tasks on client initialization
func (m *AccessKey) RegisterTasks() error {
	return nil
}

// Migrate model specific migration on startup
func (m *AccessKey) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableAccessKeys), metadataField)
}
