package bux

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bc"
)

// BlockHeader is an object representing the BitCoin block header
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type BlockHeader struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	ID                string         `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the block hash" bson:"_id"`
	Height            uint32         `json:"height" toml:"height" yaml:"height" gorm:"<-create;uniqueIndex;comment:This is the block height" bson:"height"`
	Time              uint32         `json:"time" toml:"time" yaml:"time" gorm:"<-create;index;comment:This is the time the block was mined" bson:"time"`
	Nonce             uint32         `json:"nonce" toml:"nonce" yaml:"nonce" gorm:"<-create;comment:This is the nonce" bson:"nonce"`
	Version           uint32         `json:"version" toml:"version" yaml:"version" gorm:"<-create;comment:This is the version" bson:"version"`
	HashPreviousBlock string         `json:"hash_previous_block" toml:"hash_previous_block" yaml:"hash_previous_block" gorm:"<-:create;type:char(64);index;comment:This is the hash of the previous block" bson:"hash_previous_block"`
	HashMerkleRoot    string         `json:"hash_merkle_root" toml:"hash_merkle_root" yaml:"hash_merkle_root" gorm:"<-;type:char(64);index;comment:This is the hash of the merkle root" bson:"hash_merkle_root"`
	Bits              string         `json:"bits" toml:"bits" yaml:"bits" gorm:"<-:create;comment:This is the block difficulty" bson:"bits"`
	Synced            utils.NullTime `json:"synced" toml:"synced" yaml:"synced" gorm:"type:timestamp;index;comment:This is when the block was last synced to the bux server" bson:"synced,omitempty"`
}

// newBlockHeader will start a new block header model
func newBlockHeader(hash string, blockHeader bc.BlockHeader, opts ...ModelOps) (bh *BlockHeader) {

	// Create a new model
	bh = &BlockHeader{
		ID:    hash,
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
	return m.ID
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
	return m.ID
}

// getBlockHeaderByHeight will get the block header given by height
func getBlockHeaderByHeight(ctx context.Context, height uint32, opts ...ModelOps) (*BlockHeader, error) {

	// Construct an empty model
	blockHeader := &BlockHeader{
		Height: height,
		Model:  *NewBaseModel(ModelDestination, opts...),
	}

	// Get the record
	if err := Get(ctx, blockHeader, nil, true, defaultDatabaseReadTimeout); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return blockHeader, nil
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *BlockHeader) BeforeCreating(_ context.Context) error {

	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	// Test for required field(s)
	if len(m.ID) == 0 {
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
	// import all previous block headers from file
	blockHeadersFile := m.Client().ImportBlockHeadersFromURL()
	if blockHeadersFile != "" {
		go func() {
			ctx := context.Background()
			// check whether we have block header 0, then we do not import again
			blockHeader0, err := getBlockHeaderByHeight(ctx, 0, m.Client().DefaultModelOptions()...)
			if err != nil {
				m.Client().Logger().Error(ctx, err.Error())
			} else {
				if blockHeader0 == nil {
					// import block headers in the background
					m.Client().Logger().Info(ctx, "importing block headers into database")
					if err = m.importBlockHeaders(ctx, client, blockHeadersFile); err != nil {
						m.Client().Logger().Error(ctx, err.Error())
					} else {
						m.Client().Logger().Info(ctx, "successfully imported all block headers into database")
					}
				}
			}
		}()
	}

	return nil
}

// importBlockHeaders will import the block headers from a file
func (m *BlockHeader) importBlockHeaders(ctx context.Context, client datastore.ClientInterface,
	blockHeadersFile string) error {

	file, err := ioutil.TempFile("", "blocks_bux.tsv")
	if err != nil {
		return err
	}
	defer func() {
		if err = os.Remove(file.Name()); err != nil {
			m.Client().Logger().Error(ctx, err.Error())
		}
	}()

	if err = utils.DownloadAndUnzipFile(
		ctx, m.Client().HTTPClient(), file, blockHeadersFile,
	); err != nil {
		return err
	}

	blockFile := file.Name()

	/* local file import
	var err error
	pwd, _ := os.Getwd()
	blockFile := pwd + "/blocks/blocks_bux.tsv"
	*/

	batchSize := 1000
	if m.Client().Datastore().Engine() == datastore.MongoDB {
		batchSize = 10000
	}
	models := make([]*BlockHeader, 0)
	count := 0
	readModel := func(model *BlockHeader) error {
		count++

		models = append(models, model)

		if count%batchSize == 0 {
			// insert in batches of batchSize
			if err = client.CreateInBatches(ctx, models, batchSize); err != nil {
				return err
			}
			// reset models
			models = make([]*BlockHeader, 0)
		}
		return nil
	}

	// accumulate the models into a slice
	if err = m.importCSVFile(ctx, blockFile, readModel); errors.Is(err, io.EOF) {
		if count%batchSize != 0 {
			// remaining batch
			return client.CreateInBatches(ctx, models, batchSize)
		}
		return nil
	}
	return err
}

// importCSVFile will import the block headers from a given CSV file
func (m *BlockHeader) importCSVFile(ctx context.Context, blockFile string,
	readModel func(model *BlockHeader) error) error {

	CSVFile, err := os.Open(blockFile) //nolint:gosec // file only added by administrator via config
	if err != nil {
		return err
	}
	defer func() {
		if err = CSVFile.Close(); err != nil {
			m.Client().Logger().Error(ctx, err.Error())
		}
	}()

	reader := csv.NewReader(CSVFile)
	reader.Comma = '\t'             // It's a tab-delimited file
	reader.FieldsPerRecord = 0      // -1 is variable #, 0 is [0]th line's #
	reader.LazyQuotes = true        // Some fields are like \t"F" ST.\t
	reader.TrimLeadingSpace = false // Keep the fields' whitespace how it is

	// read first line - HEADER
	if _, err = reader.Read(); err != nil {
		return err
	}

	// Read all rows
	for {
		var row []string
		if row, err = reader.Read(); err != nil {
			return err
		}

		var parsedInt uint64
		if parsedInt, err = strconv.ParseUint(row[1], 10, 32); err != nil {
			return err
		}

		height := uint32(parsedInt)

		if parsedInt, err = strconv.ParseUint(row[3], 10, 32); err != nil {
			return err
		}

		nonce := uint32(parsedInt)

		if parsedInt, err = strconv.ParseUint(row[4], 10, 32); err != nil {
			return err
		}
		ver := uint32(parsedInt)
		if parsedInt, err = strconv.ParseUint(row[7], 10, 32); err != nil {
			return err
		}
		bits := parsedInt

		var timeField time.Time
		if timeField, err = time.Parse("2006-01-02 15:04:05", row[2]); err != nil {
			return err
		}

		var syncedTime time.Time
		if syncedTime, err = time.Parse("2006-01-02 15:04:05", row[8]); err != nil {
			return err
		}

		// todo: use a function like newBlockHeader? vs making a struct
		model := &BlockHeader{
			Bits:              strconv.FormatUint(bits, 16),
			HashMerkleRoot:    row[6],
			HashPreviousBlock: row[5],
			Height:            height,
			ID:                row[0],
			Nonce:             nonce,
			Synced:            utils.NullTime{NullTime: sql.NullTime{Valid: true, Time: syncedTime}},
			Time:              uint32(timeField.Unix()),
			Version:           ver,
		}
		model.Model.CreatedAt = time.Now()

		// call the readModel callback function to add the model to the database
		if err = readModel(model); err != nil {
			return err
		}
	}
}
