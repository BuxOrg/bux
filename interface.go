package bux

import (
	"context"
	"net/http"
	"time"

	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/chainstate"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/logger"
	"github.com/BuxOrg/bux/notifications"
	"github.com/BuxOrg/bux/taskmanager"
	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bc"
	"github.com/tonicpow/go-paymail"
)

// AccessKeyService is the access key actions
type AccessKeyService interface {
	GetAccessKey(ctx context.Context, xPubID, pubAccessKey string) (*AccessKey, error)
	GetAccessKeys(ctx context.Context, xPubID string, metadata *Metadata, conditions *map[string]interface{},
		queryParams *datastore.QueryParams, opts ...ModelOps) ([]*AccessKey, error)
	NewAccessKey(ctx context.Context, rawXpubKey string, opts ...ModelOps) (*AccessKey, error)
	RevokeAccessKey(ctx context.Context, rawXpubKey, id string, opts ...ModelOps) (*AccessKey, error)
}

// TransactionService is the transaction actions
type TransactionService interface {
	GetTransaction(ctx context.Context, xPubID, txID string) (*Transaction, error)
	GetTransactions(ctx context.Context, xPubID string, metadata *Metadata, conditions *map[string]interface{},
		queryParams *datastore.QueryParams) ([]*Transaction, error)
	NewTransaction(ctx context.Context, rawXpubKey string, config *TransactionConfig,
		opts ...ModelOps) (*DraftTransaction, error)
	RecordTransaction(ctx context.Context, xPubKey, txHex, draftID string,
		opts ...ModelOps) (*Transaction, error)
	RecordMonitoredTransaction(ctx context.Context, txHex string, opts ...ModelOps) (*Transaction, error)
	UpdateTransactionMetadata(ctx context.Context, xPubID, id string, metadata Metadata) (*Transaction, error)
}

// BlockHeaderService is the block header actions
type BlockHeaderService interface {
	RecordBlockHeader(ctx context.Context, hash string, bh bc.BlockHeader, opts ...ModelOps) (*BlockHeader, error)
}

// DestinationService is the destination actions
type DestinationService interface {
	GetDestinationByID(ctx context.Context, xPubID, id string) (*Destination, error)
	GetDestinationByAddress(ctx context.Context, xPubID, address string) (*Destination, error)
	GetDestinationByLockingScript(ctx context.Context, xPubID, lockingScript string) (*Destination, error)
	GetDestinations(ctx context.Context, xPubID string, usingMetadata *Metadata, conditions *map[string]interface{},
		queryParams *datastore.QueryParams) ([]*Destination, error)
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
	GetUtxos(ctx context.Context, xPubID string, metadata *Metadata, conditions *map[string]interface{},
		queryParams *datastore.QueryParams) ([]*Utxo, error)
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
	GetPaymailAddress(ctx context.Context, address string, opts ...ModelOps) (*PaymailAddress, error)
	GetPaymailAddresses(ctx context.Context, metadataConditions *Metadata, conditions *map[string]interface{},
		queryParams *datastore.QueryParams) ([]*PaymailAddress, error)
	GetPaymailAddressesByXPubID(ctx context.Context, xPubID string, metadataConditions *Metadata,
		conditions *map[string]interface{}, queryParams *datastore.QueryParams) ([]*PaymailAddress, error)
	NewPaymailAddress(ctx context.Context, key, address, publicName, avatar string, opts ...ModelOps) (*PaymailAddress, error)
	DeletePaymailAddress(ctx context.Context, address string, opts ...ModelOps) error
	UpdatePaymailAddress(ctx context.Context, address, publicName, avatar string,
		opts ...ModelOps) (*PaymailAddress, error)
	UpdatePaymailAddressMetadata(ctx context.Context, address string,
		metadata Metadata, opts ...ModelOps) (*PaymailAddress, error)
}

// HTTPInterface is the HTTP client interface
type HTTPInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientServices is the client related services
type ClientServices interface {
	Cachestore() cachestore.ClientInterface
	Chainstate() chainstate.ClientInterface
	Datastore() datastore.ClientInterface
	HTTPClient() HTTPInterface
	Logger() logger.Interface
	Notifications() notifications.ClientInterface
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
	BlockHeaderService
	UTXOService
	XPubService
	AddModels(ctx context.Context, autoMigrate bool, models ...interface{}) error
	AuthenticateRequest(ctx context.Context, req *http.Request, adminXPubs []string,
		adminRequired, requireSigning, signingDisabled bool) (*http.Request, error)
	Close(ctx context.Context) error
	Debug(on bool)
	DefaultModelOptions(opts ...ModelOps) []ModelOps
	DefaultSyncConfig() *SyncConfig
	EnableNewRelic()
	GetFeeUnit(_ context.Context, _ string) *utils.FeeUnit
	GetOrStartTxn(ctx context.Context, name string) context.Context
	GetPaymailConfig() *PaymailServerOptions
	GetTaskPeriod(name string) time.Duration
	ImportBlockHeadersFromURL() string
	IsDebug() bool
	IsITCEnabled() bool
	IsIUCEnabled() bool
	IsNewRelicEnabled() bool
	ModifyTaskPeriod(name string, period time.Duration) error
	SetNotificationsClient(notifications.ClientInterface)
	UserAgent() string
	Version() string
}
