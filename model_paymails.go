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
	ID              string `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique paymail record id" bson:"_id"`                                                         // Unique identifier
	Alias           string `json:"alias" toml:"alias" yaml:"alias" gorm:"<-;type:varchar(64);comment:This is alias@" bson:"alias"`                                                                                   // Alias part of the paymail
	Domain          string `json:"domain" toml:"domain" yaml:"domain" gorm:"<-;type:varchar(255);comment:This is @domain.com" bson:"domain"`                                                                         // Domain of the paymail
	Username        string `json:"username" toml:"username" yaml:"username" gorm:"<-;type:varchar(255);uniqueIndex;comment:This is username" bson:"username"`                                                        // Full username
	Avatar          string `json:"avatar" toml:"avatar" yaml:"avatar" gorm:"<-;type:text;comment:This is avatar url" bson:"avatar"`                                                                                  // This is the url of the user (public profile)
	ExternalXPubKey string `json:"external_xpub_key" toml:"external_xpub_key" yaml:"external_xpub_key" gorm:"<-:create;type:varchar(111);index;comment:This is full xPub for external use" bson:"external_xpub_key"` // PublicKey hex encoded
	NextIdentityNum uint32 `json:"next_identity_num" toml:"next_identity_num" yaml:"next_identity_num" gorm:"<-;type:int;comment:The next index number for the identity xPub derivation" bson:"next_identity_num"`
	XPubID          string `json:"xpub_id" toml:"xpub_id" yaml:"xpub_id" gorm:"<-:create;type:char(64);index;comment:This is the related xPub" bson:"xpub_id"` // Related xPub ID
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
		Model:  *NewBaseModel(ModelPaymail, opts...),
	}

	// Set the xPub if detected
	if len(p.rawXpubKey) > 0 {
		_ = p.setXPub(p.rawXpubKey)
		p.XPubID = utils.Hash(p.rawXpubKey)
	}
	return p
}

// getPaymail will get the paymail with the given conditions
func getPaymail(ctx context.Context, address string, opts ...ModelOps) (*PaymailAddress, error) {

	// Get the record
	paymailAddress := newPaymail(address, opts...)
	paymailAddress.ID = ""
	conditions := map[string]interface{}{
		"alias":  paymailAddress.Alias,
		"domain": paymailAddress.Domain,
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

// getPaymailByID will get the paymail with the given ID
func getPaymailByID(ctx context.Context, id string, opts ...ModelOps) (*PaymailAddress, error) {

	// Get the record
	paymailAddress := &PaymailAddress{
		ID:    id,
		Model: *NewBaseModel(ModelPaymail, opts...),
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

// setXPub will set the "ExternalXPubKey" given the raw xPub
func (m *PaymailAddress) setXPub(rawXpubKey string) error {
	xPub, err := bitcoin.GetHDKeyFromExtendedPublicKey(rawXpubKey)
	if err != nil {
		return err
	}

	paymailExternalKey, err := bitcoin.GetHDKeyChild(
		xPub, utils.ChainExternal,
	)
	if err != nil {
		return err
	}

	m.ExternalXPubKey = paymailExternalKey.String()
	return nil
}

// GetIdentityXpub will get the identity related to the xPub
func (m *PaymailAddress) GetIdentityXpub() (*bip32.ExtendedKey, error) {
	xPub, err := bitcoin.GetHDKeyFromExtendedPublicKey(m.ExternalXPubKey)
	if err != nil {
		return nil, err
	}
	return bitcoin.GetHDKeyChild(
		xPub, uint32(utils.MaxInt32),
	)
}

// GetModelName returns the model name
func (m *PaymailAddress) GetModelName() string {
	return ModelPaymail.String()
}

// GetModelTableName returns the model db table name
func (m *PaymailAddress) GetModelTableName() string {
	return tablePaymails
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

	if _, err = utils.ValidateXPub(m.ExternalXPubKey); err != nil {
		return
	}

	if m.ID == "" {
		return ErrMissingPaymailID
	}

	if len(m.Alias) == 0 {
		return ErrMissingPaymailAddress
	}

	if len(m.Domain) == 0 {
		return ErrMissingPaymailDomain
	}

	if len(m.ExternalXPubKey) == 0 {
		return ErrMissingPaymailExternalXPub
	}

	if len(m.XPubID) == 0 {
		return ErrMissingPaymailXPubID
	}

	m.DebugLog("end: " + m.Name() + " BeforeCreating hook")
	return
}

// Migrate model specific migration on startup
func (m *PaymailAddress) Migrate(client datastore.ClientInterface) error {

	tableName := client.GetTableName(tablePaymails)
	if client.Engine() == datastore.MySQL {
		if err := m.migrateMySQL(client, tableName); err != nil {
			return err
		}
	} else if client.Engine() == datastore.PostgreSQL {
		if err := m.migratePostgreSQL(client, tableName); err != nil {
			return err
		}
	}

	return client.IndexMetadata(client.GetTableName(tablePaymails), MetadataField)
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
