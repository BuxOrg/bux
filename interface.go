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

// TransactionService is the transaction related requests
type TransactionService interface {
	GetTransaction(ctx context.Context, rawXpubKey, txID string) (*Transaction, error)
	GetTransactions(ctx context.Context, rawXpubKey string, metadata *Metadata, conditions *map[string]interface{}) ([]*Transaction, error)
	NewTransaction(ctx context.Context, rawXpubKey string, config *TransactionConfig,
		metadata map[string]interface{}, opts ...ModelOps) (*DraftTransaction, error)
	RecordTransaction(ctx context.Context, xPubKey, txHex, draftID string,
		opts ...ModelOps) (*Transaction, error)
}

// DestinationService is the destination related requests
type DestinationService interface {
	GetDestinationByAddress(ctx context.Context, xPubKey, address string) (*Destination, error)
	GetDestinationByLockingScript(ctx context.Context, xPubKey, lockingScript string) (*Destination, error)
	GetDestinations(ctx context.Context, xPubKey string, usingMetadata *Metadata) ([]*Destination, error)
	NewDestination(ctx context.Context, xPubKey string, chain uint32, destinationType string,
		metadata *map[string]interface{}) (*Destination, error)
	NewDestinationForLockingScript(ctx context.Context, xPubID, lockingScript, destinationType string,
		metadata map[string]interface{}) (*Destination, error)
}

// UTXOService is the utxo related requests
type UTXOService interface {
	GetUtxo(ctx context.Context, xPubKey, txID string, outputIndex uint32) (*Utxo, error)
	GetUtxos(ctx context.Context, xPubKey string) ([]*Utxo, error)
}

// XPubService is the xPub related requests
type XPubService interface {
	GetXpub(ctx context.Context, xPubKey string) (*Xpub, error)
	GetXpubByID(ctx context.Context, xPubID string) (*Xpub, error)
	NewXpub(ctx context.Context, xPubKey string, opts ...ModelOps) (*Xpub, error)
}

// ClientInterface is the client (bux engine) interface
type ClientInterface interface {
	DestinationService
	TransactionService
	UTXOService
	XPubService
	AddModels(ctx context.Context, autoMigrate bool, models ...interface{}) error
	AuthenticateRequest(ctx context.Context, req *http.Request, adminXPubs []string, adminRequired, requireSigning, signingDisabled bool) (*http.Request, error)
	Cachestore() cachestore.ClientInterface
	Chainstate() chainstate.ClientInterface
	Close(ctx context.Context) error
	Datastore() datastore.ClientInterface
	Debug(on bool)
	DefaultModelOptions(opts ...ModelOps) []ModelOps
	EnableNewRelic()
	GetFeeUnit(_ context.Context, _ string) *utils.FeeUnit
	GetOrStartTxn(ctx context.Context, name string) context.Context
	GetTaskPeriod(name string) time.Duration
	IsDebug() bool
	IsITCEnabled() bool
	IsInputUtxoCheckEnabled() bool
	IsNewRelicEnabled() bool
	Logger() logger.Interface
	ModifyPaymailConfig(config *server.Configuration, defaultFromPaymail, defaultNote string)
	ModifyTaskPeriod(name string, period time.Duration) error
	PaymailClient() paymail.ClientInterface
	PaymailServerConfig() *paymailServerOptions
	Taskmanager() taskmanager.ClientInterface
	UserAgent() string
	Version() string
}
