package bux

import (
	"context"
	"errors"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bip32"
	"github.com/tonicpow/go-paymail"
)

// PaymailAddress is an "external model example" - this model is not part of the standard models loaded and runtime
//
// This model must be included at runtime via WithAutoMigrate() etc...
//
// Gorm related models & indexes: https://gorm.io/docs/models.html - https://gorm.io/docs/indexes.html
type PaymailAddress struct {
	// Base model
	Model `bson:",inline"`

	// Model specific fields
	ID              string `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique paymail record id" bson:"_id"`                                                                              // Unique identifier
	XpubID          string `json:"xpub_id" toml:"xpub_id" yaml:"xpub_id" gorm:"<-:create;type:char(64);index;comment:This is the related xPub" bson:"xpub_id"`                                                                            // Related xPub ID
	Alias           string `json:"alias" toml:"alias" yaml:"alias" gorm:"<-;type:varchar(64);comment:This is alias@" bson:"alias"`                                                                                                        // Alias part of the paymail
	Domain          string `json:"domain" toml:"domain" yaml:"domain" gorm:"<-;type:varchar(255);comment:This is @domain.com" bson:"domain"`                                                                                              // Domain of the paymail
	PublicName      string `json:"public_name" toml:"public_name" yaml:"public_name" gorm:"<-;type:varchar(255);comment:This is public name for public profile" bson:"public_name,omitempty"`                                             // Full username
	Avatar          string `json:"avatar" toml:"avatar" yaml:"avatar" gorm:"<-;type:text;comment:This is avatar url" bson:"avatar"`                                                                                                       // This is the url of the user (public profile)
	ExternalXpubKey string `json:"external_xpub_key" toml:"external_xpub_key" yaml:"external_xpub_key" gorm:"<-:create;type:varchar(512);index;comment:This is full xPub for external use, encryption optional" bson:"external_xpub_key"` // PublicKey hex encoded

	// Private fields
	externalXpubKeyDecrypted string
}

// newPaymail create new paymail model
func newPaymail(paymailAddress string, opts ...ModelOps) *PaymailAddress {

	// Standardize and sanitize!
	alias, domain, _ := paymail.SanitizePaymail(paymailAddress)
	id, _ := utils.RandomHex(32)
	p := &PaymailAddress{
		Alias:  alias,
		Domain: domain,
		ID:     id,
		Model:  *NewBaseModel(ModelPaymailAddress, opts...),
	}

	// Set the xPub information if found
	if len(p.rawXpubKey) > 0 {
		_ = p.setXPub()
	}
	return p
}

// getPaymail will get the paymail with the given conditions
func getPaymail(ctx context.Context, address string, opts ...ModelOps) (*PaymailAddress, error) {

	// Get the record
	paymailAddress := newPaymail(address, opts...)
	paymailAddress.ID = ""
	conditions := map[string]interface{}{
		aliasField:  paymailAddress.Alias,
		domainField: paymailAddress.Domain,
	}

	if err := Get(
		ctx, paymailAddress, conditions, false, defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return paymailAddress, nil
}

// getPaymails will get all the paymails with the given conditions
func getPaymails(ctx context.Context, metadata *Metadata, conditions *map[string]interface{},
	queryParams *datastore.QueryParams, opts ...ModelOps) ([]*PaymailAddress, error) {

	var models []PaymailAddress
	dbConditions := map[string]interface{}{}

	if metadata != nil {
		dbConditions[metadataField] = metadata
	}

	if conditions != nil && len(*conditions) > 0 {
		and := make([]map[string]interface{}, 0)
		if _, ok := dbConditions["$and"]; ok {
			and = dbConditions["$and"].([]map[string]interface{})
		}
		and = append(and, *conditions)
		dbConditions["$and"] = and
	}

	// Get the records
	if err := getModels(
		ctx, NewBaseModel(ModelNameEmpty, opts...).Client().Datastore(),
		&models, dbConditions, queryParams, defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	// Loop and enrich
	destinations := make([]*PaymailAddress, 0)
	for index := range models {
		models[index].enrich(ModelDestination, opts...)
		destinations = append(destinations, &models[index])
	}

	return destinations, nil
}

// getPaymailByID will get the paymail with the given ID
func getPaymailByID(ctx context.Context, id string, opts ...ModelOps) (*PaymailAddress, error) {

	// Get the record
	paymailAddress := &PaymailAddress{
		ID:    id,
		Model: *NewBaseModel(ModelPaymailAddress, opts...),
	}
	if err := Get(
		ctx, paymailAddress, nil, false, defaultDatabaseReadTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return paymailAddress, nil
}

// setXPub will set the "ExternalXPubKey" given the raw xPub and xPubID
// encrypted with the given encryption key (if a key is set)
func (m *PaymailAddress) setXPub() error {

	// Set the ID
	m.XpubID = utils.Hash(m.rawXpubKey)

	// Derive the public key from string
	xPub, err := bitcoin.GetHDKeyFromExtendedPublicKey(m.rawXpubKey)
	if err != nil {
		return err
	}

	// Get the external public key
	var paymailExternalKey *bip32.ExtendedKey
	if paymailExternalKey, err = bitcoin.GetHDKeyChild(
		xPub, utils.ChainExternal,
	); err != nil {
		return err
	}

	// Set the decrypted version
	m.externalXpubKeyDecrypted = paymailExternalKey.String()

	// Encrypt the xPub
	if len(m.encryptionKey) > 0 {
		m.ExternalXpubKey, err = utils.Encrypt(m.encryptionKey, m.externalXpubKeyDecrypted)
	} else {
		m.ExternalXpubKey = m.externalXpubKeyDecrypted
	}

	return err
}

// GetIdentityXpub will get the identity related to the xPub
func (m *PaymailAddress) GetIdentityXpub() (*bip32.ExtendedKey, error) {

	// Get the external xPub (to derive the identity key)
	xPub, err := m.GetExternalXpub()
	if err != nil {
		return nil, err
	}

	// Get the last possible key in the external key
	return bitcoin.GetHDKeyChild(
		xPub, uint32(utils.MaxInt32),
	)
}

// GetExternalXpub will get the external xPub
func (m *PaymailAddress) GetExternalXpub() (*bip32.ExtendedKey, error) {

	// Check if the xPub was encrypted
	if len(m.ExternalXpubKey) != utils.XpubKeyLength {
		var err error
		if m.externalXpubKeyDecrypted, err = utils.Decrypt(
			m.encryptionKey, m.ExternalXpubKey,
		); err != nil {
			return nil, err
		}
	} else {
		m.externalXpubKeyDecrypted = m.ExternalXpubKey
	}

	// Get the xPub
	xPub, err := bitcoin.GetHDKeyFromExtendedPublicKey(m.externalXpubKeyDecrypted)
	if err != nil {
		return nil, err
	}
	return xPub, nil
}

// GetModelName returns the model name
func (m *PaymailAddress) GetModelName() string {
	return ModelPaymailAddress.String()
}

// GetModelTableName returns the model db table name
func (m *PaymailAddress) GetModelTableName() string {
	return tablePaymailAddresses
}

// Save the model
func (m *PaymailAddress) Save(ctx context.Context) (err error) {
	return Save(ctx, m)
}

// GetID will get the ID
func (m *PaymailAddress) GetID() string {
	return m.ID
}

// BeforeCreating is called before the model is saved to the DB
func (m *PaymailAddress) BeforeCreating(_ context.Context) (err error) {
	m.DebugLog("starting: " + m.Name() + " BeforeCreating hook...")

	if m.ID == "" {
		return ErrMissingPaymailID
	}

	if len(m.Alias) == 0 {
		return ErrMissingPaymailAddress
	}

	if len(m.Domain) == 0 {
		return ErrMissingPaymailDomain
	}

	if len(m.ExternalXpubKey) == 0 {
		return ErrMissingPaymailExternalXPub
	} else if len(m.externalXpubKeyDecrypted) > 0 {
		if _, err = utils.ValidateXPub(m.externalXpubKeyDecrypted); err != nil {
			return
		}
	}

	if len(m.XpubID) == 0 {
		return ErrMissingPaymailXPubID
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return
}

// Migrate model specific migration on startup
func (m *PaymailAddress) Migrate(client datastore.ClientInterface) error {

	tableName := client.GetTableName(tablePaymailAddresses)
	if client.Engine() == datastore.MySQL {
		if err := m.migrateMySQL(client, tableName); err != nil {
			return err
		}
	} else if client.Engine() == datastore.PostgreSQL {
		if err := m.migratePostgreSQL(client, tableName); err != nil {
			return err
		}
	}

	return client.IndexMetadata(client.GetTableName(tablePaymailAddresses), MetadataField)
}

// migratePostgreSQL is specific migration SQL for Postgresql
func (m *PaymailAddress) migratePostgreSQL(client datastore.ClientInterface, tableName string) error {
	tx := client.Execute(`CREATE UNIQUE INDEX IF NOT EXISTS "idx_paymail_address" ON "` + tableName + `" ("alias","domain")`)
	return tx.Error
}

// migrateMySQL is specific migration SQL for MySQL
func (m *PaymailAddress) migrateMySQL(client datastore.ClientInterface, tableName string) error {
	tx := client.Execute("CREATE UNIQUE INDEX idx_paymail_address ON `" + tableName + "` (alias, domain)")
	return tx.Error
}
