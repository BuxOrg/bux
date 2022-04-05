package bux

import (
	"context"
	"net/http"
	"time"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/utils"
	"github.com/tonicpow/go-paymail"
	"github.com/tonicpow/go-paymail/server"
	"gorm.io/gorm/logger"
)

// AccessKeyService is the access key actions
type AccessKeyService interface {
	GetAccessKey(ctx context.Context, xPubID, pubAccessKey string) (*AccessKey, error)
	GetAccessKeys(ctx context.Context, xPubID string, metadata *Metadata, opts ...ModelOps) ([]*AccessKey, error)
	NewAccessKey(ctx context.Context, rawXpubKey string, opts ...ModelOps) (*AccessKey, error)
	RevokeAccessKey(ctx context.Context, rawXpubKey, id string, opts ...ModelOps) (*AccessKey, error)
}

// TransactionService is the transaction actions
type TransactionService interface {
	GetTransaction(ctx context.Context, xPubID, txID string) (*Transaction, error)
	GetTransactions(ctx context.Context, xPubID string, metadata *Metadata, conditions *map[string]interface{}) ([]*Transaction, error)
	NewTransaction(ctx context.Context, rawXpubKey string, config *TransactionConfig,
		opts ...ModelOps) (*DraftTransaction, error)
	RecordTransaction(ctx context.Context, xPubKey, txHex, draftID string,
		opts ...ModelOps) (*Transaction, error)
	UpdateTransactionMetadata(ctx context.Context, xPubID, id string, metadata Metadata) (*Transaction, error)
}

// DestinationService is the destination actions
type DestinationService interface {
	GetDestinationByID(ctx context.Context, xPubID, id string) (*Destination, error)
	GetDestinationByAddress(ctx context.Context, xPubID, address string) (*Destination, error)
	GetDestinationByLockingScript(ctx context.Context, xPubID, lockingScript string) (*Destination, error)
	GetDestinations(ctx context.Context, xPubID string, usingMetadata *Metadata) ([]*Destination, error)
	NewDestination(ctx context.Context, xPubKey string, chain uint32, destinationType string, monitor bool,
		opts ...ModelOps) (*Destination, error)
	NewDestinationForLockingScript(ctx context.Context, xPubID, lockingScript string, monitor bool,
		opts ...ModelOps) (*Destination, error)
	UpdateDestinationMetadataByID(ctx context.Context, xPubID, id string, metadata Metadata) (*Destination, error)
	UpdateDestinationMetadataByLockingScript(ctx context.Context, xPubID, lockingScript string, metadata Metadata) (*Destination, error)
	UpdateDestinationMetadataByAddress(ctx context.Context, xPubID, address string, metadata Metadata) (*Destination, error)
}

// UTXOService is the utxo actions
type UTXOService interface {
	GetUtxo(ctx context.Context, xPubKey, txID string, outputIndex uint32) (*Utxo, error)
	GetUtxos(ctx context.Context, xPubKey string) ([]*Utxo, error)
}

// XPubService is the xPub actions
type XPubService interface {
	GetXpub(ctx context.Context, xPubKey string) (*Xpub, error)
	GetXpubByID(ctx context.Context, xPubID string) (*Xpub, error)
	ImportXpub(ctx context.Context, xPubKey string, opts ...ModelOps) (*ImportResults, error)
	NewXpub(ctx context.Context, xPubKey string, opts ...ModelOps) (*Xpub, error)
	UpdateXpubMetadata(ctx context.Context, xPubID string, metadata Metadata) (*Xpub, error)
}

// PaymailService is the paymail actions
type PaymailService interface {
	NewPaymailAddress(ctx context.Context, key, address string, opts ...ModelOps) (*PaymailAddress, error)
	DeletePaymailAddress(ctx context.Context, address string, opts ...ModelOps) error
	UpdatePaymailAddressMetadata(ctx context.Context, address string,
		metadata Metadata, opts ...ModelOps) (*PaymailAddress, error)
}

// ClientServices is the client related services
type ClientServices interface {
	Cachestore() cachestore.ClientInterface
	Chainstate() chainstate.ClientInterface
	Datastore() datastore.ClientInterface
	Logger() logger.Interface
	PaymailClient() paymail.ClientInterface
	Taskmanager() taskmanager.ClientInterface
}

// ClientInterface is the client (bux engine) interface comprised of all services
type ClientInterface interface {
	AccessKeyService
	ClientServices
	DestinationService
	PaymailService
	TransactionService
	UTXOService
	XPubService
	AddModels(ctx context.Context, autoMigrate bool, models ...interface{}) error
	AuthenticateRequest(ctx context.Context, req *http.Request, adminXPubs []string, adminRequired, requireSigning, signingDisabled bool) (*http.Request, error)
	Close(ctx context.Context) error
	Debug(on bool)
	DefaultModelOptions(opts ...ModelOps) []ModelOps
	EnableNewRelic()
	GetFeeUnit(_ context.Context, _ string) *utils.FeeUnit
	GetOrStartTxn(ctx context.Context, name string) context.Context
	GetTaskPeriod(name string) time.Duration
	IsDebug() bool
	IsITCEnabled() bool
	IsIUCEnabled() bool
	IsNewRelicEnabled() bool
	ModifyPaymailConfig(config *server.Configuration, defaultFromPaymail, defaultNote string)
	ModifyTaskPeriod(name string, period time.Duration) error
	PaymailServerConfig() *PaymailServerOptions
	UserAgent() string
	Version() string
}
