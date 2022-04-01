package bux

import (
	"context"
	"encoding/hex"

	"github.com/BuxOrg/bux/datastore"
	"github.com/libsv/go-bc"
)

// BlockHeader is an object representing the BitCoin block header
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type BlockHeader struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	Hash              string `json:"hash" toml:"hash" yaml:"hash" gorm:"<-:create;type:char(64);primaryKey;comment:This is the block header" bson:"hash"`
	Height            uint32 `json:"height" toml:"height" yaml:"height" gorm:"<-create;type:int;uniqueIndex;comment:This is the block height" bson:"height,omitempty"`
	Time              uint32 `json:"time" toml:"time" yaml:"time" gorm:"<-create;type:int;index;comment:This is the time the block was mined" bson:"time,omitempty"`
	Nonce             uint32 `json:"nonce" toml:"nonce" yaml:"nonce" gorm:"<-create;type:int;comment:This is the nonce" bson:"nonce,omitempty"`
	Version           uint32 `json:"version" toml:"version" yaml:"version" gorm:"<-create;type:int;comment:This is the version" bson:"version,omitempty"`
	HashPreviousBlock string `json:"hash_previous_block" toml:"hash_previous_block" yaml:"hash_previous_block" gorm:"<-:create;type:text;index;comment:This is the hash of the previous block" bson:"hash_previous_block"`
	HashMerkleRoot    string `json:"hash_merkle_root" toml:"hash_merkle_root" yaml:"hash_merkle_root" gorm:"<-;type:text;index;comment:This is the hash of the merkle root" bson:"hash_merkle_root"`
	Bits              string `json:"bits" toml:"bits" yaml:"bits" gorm:"<-:create;type:text;comment:This is the block difficulty" bson:"bits"`
}

// newBlockHeader will start a new transaction model
func newBlockHeader(hash string, blockHeader bc.BlockHeader, opts ...ModelOps) (bh *BlockHeader) {

	// Create a new model
	bh = &BlockHeader{
		Hash:  hash,
		Model: *NewBaseModel(ModelBlockHeader, opts...),
	}

	// Set header info
	bh.setHeaderInfo(blockHeader)
	return
}

// GetModelName will get the name of the current model
func (m *BlockHeader) GetModelName() string {
	return ModelBlockHeader.String()
}

// GetModelTableName will get the db table name of the current model
func (m *BlockHeader) GetModelTableName() string {
	return tableBlockHeaders
}

// Save will Save the model into the Datastore
func (m *BlockHeader) Save(ctx context.Context) (err error) {
	return Save(ctx, m)
}

// GetHash will get the hash of the block header
func (m *BlockHeader) GetHash() string {
	return m.Hash
}

// setHeaderInfo will set the block header info from a bc.BlockHeader
func (m *BlockHeader) setHeaderInfo(bh bc.BlockHeader) {
	m.Bits = hex.EncodeToString(bh.Bits)
	m.HashMerkleRoot = hex.EncodeToString(bh.HashMerkleRoot)
	m.HashPreviousBlock = hex.EncodeToString(bh.HashPrevBlock)
	m.Nonce = bh.Nonce
	m.Time = bh.Time
	m.Version = bh.Version
}

// GetID will return the id of the field (hash)
func (m *BlockHeader) GetID() string {
	return m.Hash
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *BlockHeader) BeforeCreating(_ context.Context) error {

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Test for required field(s)
	if len(m.Hash) == 0 {
		return ErrMissingFieldHash
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *BlockHeader) AfterCreated(_ context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

// Display filter the model for display
func (m *BlockHeader) Display() interface{} {
	return m
}

// Migrate model specific migration on startup
func (m *BlockHeader) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tableBlockHeaders), metadataField)
}
