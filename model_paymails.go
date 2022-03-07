package bux

import (
	"context"
	"errors"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
	"github.com/tonicpow/go-paymail"
)

// PaymailAddress is an external bux model example
type PaymailAddress struct {
	Model           `bson:",inline"` // Base bux model
	ID              string           `json:"id" toml:"id" yaml:"id" gorm:"<-:create;type:char(64);primaryKey;comment:This is the unique paymail record id" bson:"_id"`                                                         // Unique identifier
	Alias           string           `json:"alias" toml:"alias" yaml:"alias" gorm:"<-;type:varchar(64);uniqueIndex:idx_paymail_address;comment:This is alias@" bson:"alias"`                                                   // Alias part of the paymail
	Domain          string           `json:"domain" toml:"domain" yaml:"domain" gorm:"<-;type:varchar(255);uniqueIndex:idx_paymail_address;comment:This is @domain.com" bson:"domain"`                                         // Domain of the paymail
	Username        string           `json:"username" toml:"username" yaml:"username" gorm:"<-;type:varchar(255);uniqueIndex;comment:This is username" bson:"username"`                                                        // Full username
	Avatar          string           `json:"avatar" toml:"avatar" yaml:"avatar" gorm:"<-;type:text;comment:This is avatar url" bson:"avatar"`                                                                                  // This is the url of the user (public profile)
	ExternalXPubKey string           `json:"external_xpub_key" toml:"external_xpub_key" yaml:"external_xpub_key" gorm:"<-:create;type:varchar(111);index;comment:This is full xPub for external use" bson:"external_xpub_key"` // PublicKey hex encoded
	XPubID          string           `json:"xpub_id" toml:"xpub_id" yaml:"xpub_id" gorm:"<-:create;type:char(64);index;comment:This is the related xPub" bson:"xpub_id"`                                                       // PublicKey hex encoded
}

const (
	defaultGetTimeout       = 10 * time.Second
	paymailRequestField     = "paymail_request"
	paymailMetadataField    = "paymail_metadata"
	paymailP2PMetadataField = "p2p_tx_metadata"
)

// NewPaymail create new paymail model
func NewPaymail(paymailAddress string, opts ...ModelOps) *PaymailAddress {

	// Standardize and sanitize!
	alias, domain, _ := paymail.SanitizePaymail(paymailAddress)
	id, _ := utils.RandomHex(32)
	return &PaymailAddress{
		Model:  *NewBaseModel(ModelPaymail, opts...),
		Alias:  alias,
		Domain: domain,
		ID:     id,
	}
}

// GetPaymail will get the paymail with the given conditions
func GetPaymail(ctx context.Context, address string, opts ...ModelOps) (*PaymailAddress, error) {

	alias, domain, _ := paymail.SanitizePaymail(address)
	// Get the record
	paymailAddress := &PaymailAddress{
		Model: *NewBaseModel(ModelXPub, opts...),
	}

	conditions := map[string]interface{}{
		"alias":  alias,
		"domain": domain,
	}

	if err := Get(
		ctx, paymailAddress, conditions, false, defaultGetTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return paymailAddress, nil
}

// GetPaymailByID will get the paymail with the given ID
func GetPaymailByID(ctx context.Context, id string, opts ...ModelOps) (*PaymailAddress, error) {

	// Get the record
	paymailAddress := &PaymailAddress{
		ID:    id,
		Model: *NewBaseModel(ModelXPub, opts...),
	}
	if err := Get(
		ctx, paymailAddress, nil, false, defaultGetTimeout,
	); err != nil {
		if errors.Is(err, datastore.ErrNoResults) {
			return nil, nil
		}
		return nil, err
	}

	return paymailAddress, nil
}

// GetModelName returns the model name
func (p *PaymailAddress) GetModelName() string {
	return ModelPaymail.String()
}

// GetModelTableName returns the model db table name
func (p *PaymailAddress) GetModelTableName() string {
	return tablePaymails
}

// Save paymail
func (p *PaymailAddress) Save(ctx context.Context) (err error) {
	return Save(ctx, p)
}

// GetID will get the ID
func (p *PaymailAddress) GetID() string {
	return p.ID
}

// BeforeCreating is called before the model is saved to the DB
func (p *PaymailAddress) BeforeCreating(_ context.Context) (err error) {
	p.DebugLog("starting: " + p.Name() + " BeforeCreating hook...")

	if _, err = utils.ValidateXPub(p.ExternalXPubKey); err != nil {
		return
	}

	if p.ID == "" {
		return ErrMissingPaymailID
	}

	if len(p.Alias) == 0 {
		return ErrMissingPaymailAddress
	}

	if len(p.Domain) == 0 {
		return ErrMissingPaymailDomain
	}

	if len(p.ExternalXPubKey) == 0 {
		return ErrMissingPaymailExternalXPub
	}

	if len(p.XPubID) == 0 {
		return ErrMissingPaymailXPubID
	}

	p.DebugLog("end: " + p.Name() + " BeforeCreating hook")
	return
}

// Migrate model specific migration
func (p *PaymailAddress) Migrate(client datastore.ClientInterface) error {
	return client.IndexMetadata(client.GetTableName(tablePaymails), MetadataField)
}
