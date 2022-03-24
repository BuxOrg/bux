package bux

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/BuxOrg/bux/datastore"
	"github.com/libsv/go-bc"
)

// BlockHeaderBase is the same fields share between multiple transaction models
type BlockHeaderBase struct {
	Hash string `json:"hash" toml:"hash" yaml:"hash" gorm:"<-:create;type:text;comment:This is the raw transaction hash" bson:"hash"`
}

// BlockHeader is an object representing the BitCoin transaction table
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type BlockHeader struct {
	// Base model
	Model `bson:",inline"`

	// Standard transaction model base fields
	BlockHeaderBase `bson:",inline"`

	// Model specific fields
	Version           uint32 `json:"version" toml:"version" yaml:"version" gorm:"<-create;type:int" bson:"version,omitempty"`
	Height            uint32 `json:"height" toml:"height" yaml:"height" gorm:"<-create;type:int" bson:"height,omitempty"`
	Time              uint32 `json:"time" toml:"time" yaml:"time" gorm:"<-create;type:int" bson:"time,omitempty"`
	Nonce             uint32 `json:"nonce" toml:"nonce" yaml:"nonce" gorm:"<-create;type:int" bson:"nonce,omitempty"`
	HashPreviousBlock string `json:"hashPreviousBlock" toml:"hashPreviousBlock" yaml:"hashPreviousBlock" gorm:"<-:create;type:text;comment:This is the raw transaction hex" bson:"hashPreviousBlock"`
	HashMerkleRoot    string `json:"hashMerkleRoot" toml:"hashMerkleRoot" yaml:"hashMerkleRoot" gorm:"<-:create;type:text;comment:This is the raw transaction hex" bson:"hashMerkleRoot"`
	Bits              string `json:"bits" toml:"bits" yaml:"bits" gorm:"<-:create;type:text;comment:This is the block difficulty" bson:"bits"`
}

// newBlockHeaderBase creates the standard transaction model base
func newBlockHeaderBase(hash string, opts ...ModelOps) *BlockHeader {
	return &BlockHeader{
		BlockHeaderBase: BlockHeaderBase{
			Hash: hash,
		},
		Model:              *NewBaseModel(ModelBlockHeader, opts...),
		Status:             statusComplete,
		blockHeaderService: blockHeadersService{},
	}
}

// newBlockHeader will start a new transaction model
func newBlockHeader(hash string, bh bc.BlockHeader, opts ...ModelOps) (nbh *BlockHeader) {
	nbh = newBlockHeaderBase(hash, opts...)

	// Set header info
	nbh.setHeaderInfo(bh)

	return
}

// getBlockHeaderByID will get the model from a given transaction ID
func getBlockHeaderByID(ctx context.Context, hash string, opts ...ModelOps) (*BlockHeader, error) {

	// Construct an empty tx
	bh := newBlockHeader("", opts...)
	bh.Hash = hash

	// Get the record
	if err := Get(ctx, bh, nil, false, defaultDatabaseReadTimeout); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return bh, nil
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
func (m *BlockHeader) setHeaderInfo(bh bc.BlockHeader) (err error) {
	m.Time = bh.Time
	m.Nonce = bh.Nonce
	m.Version = bh.Version
	m.HashPrevBlock = hex.EncodeToString(bh.HashPrevBlock)
	m.HashMerkleRoot = hex.EncodeToString(bh.HashMerkleRoot)
	m.Bits = hex.EncodeToString(bh.Bits)

	return
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *BlockHeader) BeforeCreating(ctx context.Context) error {

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Test for required field(s)
	if len(m.Hash) == 0 {
		return ErrMissingFieldHash
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *BlockHeader) AfterCreated(ctx context.Context) error {
	m.DebugLog("starting: " + m.Name() + " AfterCreated hook...")

	// Pre-build the options
	opts := m.GetOptions(false)

	// todo: run these in go routines?

	m.DebugLog("end: " + m.Name() + " AfterCreated hook")
	return nil
}

// Display filter the model for display
func (m *BlockHeader) Display() interface{} {
	return m
}

// Migrate model specific migration on startup
func (m *BlockHeader) Migrate(client datastore.ClientInterface) error {

	tableName := client.GetTableName(tableBlockHeaders)
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
func (m *BlockHeader) migratePostgreSQL(client datastore.ClientInterface, tableName string) error {

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
func (m *BlockHeader) migrateMySQL(client datastore.ClientInterface, tableName string) error {

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
			return nil // nolint: nilerr // error is not needed
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
			return nil // nolint: nilerr // error is not needed
		}
	}

	return nil
}
