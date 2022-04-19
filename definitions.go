package bux

import (
	"time"

	"github.com/BuxOrg/bux/utils"
)

// Defaults for engine functionality
const (
	databaseLongReadTimeout    = 30 * time.Second  // For all "GET" or "SELECT" methods
	defaultCacheLockTTL        = 20                // in Seconds
	defaultCacheLockTTW        = 10                // in Seconds
	defaultDatabaseReadTimeout = 20 * time.Second  // For all "GET" or "SELECT" methods
	defaultDraftTxExpiresIn    = 30 * time.Second  // Default TTL for draft transactions
	defaultHTTPTimeout         = 20 * time.Second  //
	defaultOverheadSize        = uint64(10)        // 10 bytes is the default overhead in a transaction
	defaultUserAgent           = "bux: " + version // Default user agent
	dustLimit                  = uint64(546)       // Dust limit
	mongoTestVersion           = "4.2.1"           // Mongo Testing Version
	sqliteTestVersion          = "3.37.0"          // SQLite Testing Version (dummy version for now)
	version                    = "v0.2.7"          // bux version
)

// All the base models
const (
	ModelAccessKey           ModelName = "access_key"
	ModelDestination         ModelName = "destination"
	ModelDraftTransaction    ModelName = "draft_transaction"
	ModelIncomingTransaction ModelName = "incoming_transaction"
	ModelMetadata            ModelName = "metadata"
	ModelNameEmpty           ModelName = "empty"
	ModelPaymailAddress      ModelName = "paymail_address"
	ModelSyncTransaction     ModelName = "sync_transaction"
	ModelTransaction         ModelName = "transaction"
	ModelBlockHeader         ModelName = "block_header"
	ModelUtxo                ModelName = "utxo"
	ModelXPub                ModelName = "xpub"
)

var (
	// AllModelNames is a list of all models
	AllModelNames = []ModelName{
		ModelAccessKey,
		ModelBlockHeader,
		ModelBlockHeader,
		ModelDestination,
		ModelIncomingTransaction,
		ModelMetadata,
		ModelPaymailAddress,
		ModelSyncTransaction,
		ModelTransaction,
		ModelUtxo,
		ModelXPub,
		ModelPaymailAddress,
		ModelBlockHeader,
	}
)

// Internal table names
const (
	tableAccessKeys           = "access_keys"
	tableBlockHeaders         = "block_headers"
	tableDestinations         = "destinations"
	tableDraftTransactions    = "draft_transactions"
	tableIncomingTransactions = "incoming_transactions"
	tablePaymailAddresses     = "paymail_addresses"
	tableSyncTransactions     = "sync_transactions"
	tableTransactions         = "transactions"
	tableUTXOs                = "utxos"
	tableXPubs                = "xpubs"
)

const (
	// ReferenceIDField is used for Paymail
	ReferenceIDField = "reference_id"

	// Internal field names
	aliasField           = "alias"
	broadcastStatusField = "broadcast_status"
	currentBalanceField  = "current_balance"
	domainField          = "domain"
	draftIDField         = "draft_id"
	idField              = "id"
	metadataField        = "metadata"
	nextExternalNumField = "next_external_num"
	nextInternalNumField = "next_internal_num"
	satoshisField        = "satoshis"
	spendingTxIDField    = "spending_tx_id"
	statusField          = "status"
	syncStatusField      = "sync_status"
	typeField            = "type"
	xPubIDField          = "xpub_id"
	xPubMetadataField    = "xpub_metadata"

	// Universal statuses
	statusCanceled   = "canceled"
	statusComplete   = "complete"
	statusDraft      = "draft"
	statusError      = "error"
	statusExpired    = "expired"
	statusPending    = "pending"
	statusProcessing = "processing"
	statusReady      = "ready"
	statusSkipped    = "skipped"

	// Paymail / Handles
	cacheKeyAddressResolution       = "paymail-address-resolution-"
	cacheKeyCapabilities            = "paymail-capabilities-"
	cacheTTLAddressResolution       = 2 * time.Minute
	cacheTTLCapabilities            = 60 * time.Minute
	defaultAddressResolutionPurpose = "Created with BUX: getbux.io"
	defaultSenderPaymail            = "buxorg@moneybutton.com"
	handleHandcashPrefix            = "$"
	handleMaxLength                 = 25
	handleRelayPrefix               = "1"
	p2pMetadataField                = "p2p_tx_metadata"

	// Misc
	gormTypeText = "text"
	migrateList  = "migrate"
	modelList    = "models"
)

var (
	// defaultFee is used when a fee is not provided the draft transaction
	defaultFee = &utils.FeeUnit{
		Satoshis: 1,
		Bytes:    2,
	}

	// BaseModels is the list of models for loading the engine and AutoMigration (defaults)
	BaseModels = []interface{}{

		// Base extended HD-key table
		&Xpub{
			Model: *NewBaseModel(ModelXPub),
		},

		// Access keys (extend access from xPub)
		&AccessKey{
			Model: *NewBaseModel(ModelAccessKey),
		},

		// Draft transactions are created before the final transaction is completed
		&DraftTransaction{
			Model: *NewBaseModel(ModelDraftTransaction),
		},

		// Incoming transactions (external & unknown) (related to Transaction & Draft)
		&IncomingTransaction{
			Model: *NewBaseModel(ModelIncomingTransaction),
		},

		// Finalized transactions (related to Draft)
		&Transaction{
			Model: *NewBaseModel(ModelTransaction),
		},

		// Block Headers as received by the BitCoin network
		&BlockHeader{
			Model: *NewBaseModel(ModelBlockHeader),
		},

		// Sync configuration for transactions (on-chain) (related to Transaction)
		&SyncTransaction{
			Model: *NewBaseModel(ModelSyncTransaction),
		},

		// Various types of destinations (common is: P2PKH Address)
		&Destination{
			Model: *NewBaseModel(ModelDestination),
		},

		// Unspent outputs from known transactions
		&Utxo{
			Model: *NewBaseModel(ModelUtxo),
		},

		// Paymail addresses related to XPubs (automatically added when paymail is enabled)
		/*&PaymailAddress{
			Model: *NewBaseModel(ModelPaymailAddress),
		},*/
	}
)
