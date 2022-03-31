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
	defaultOverheadSize        = uint64(10)        // 10 bytes is the default overhead in a transaction
	defaultUserAgent           = "bux: " + version // Default user agent
	dustLimit                  = uint64(546)       // Dust limit
	mongoTestVersion           = "4.2.1"           // Mongo Testing Version
	sqliteTestVersion          = "3.37.0"          // SQLite Testing Version (dummy version for now)
	version                    = "v0.2.0"          // bux version
)

// All the base models
const (
	ModelAccessKey           ModelName = "access_key"
	ModelDestination         ModelName = "destination"
	ModelDraftTransaction    ModelName = "draft_transaction"
	ModelIncomingTransaction ModelName = "incoming_transaction"
	ModelMetadata            ModelName = "metadata"
	ModelNameEmpty           ModelName = "empty"
	ModelSyncTransaction     ModelName = "sync_transaction"
	ModelTransaction         ModelName = "transaction"
	ModelBlockHeader         ModelName = "block_header"
	ModelUtxo                ModelName = "utxo"
	ModelXPub                ModelName = "xpub"
	ModelPaymail             ModelName = "paymail"
)

var (
	// AllModelNames is a list of all models
	AllModelNames = []ModelName{
		ModelAccessKey,
		ModelDestination,
		ModelIncomingTransaction,
		ModelMetadata,
		ModelSyncTransaction,
		ModelTransaction,
		ModelBlockHeader,
		ModelUtxo,
		ModelXPub,
		ModelPaymail,
		ModelBlockHeader,
	}
)

// Internal table names
const (
	tableAccessKeys           = "access_keys"
	tableDestinations         = "destinations"
	tableDraftTransactions    = "draft_transactions"
	tableIncomingTransactions = "incoming_transactions"
	tableSyncTransactions     = "sync_transactions"
	tableTransactions         = "transactions"
	tableBlockHeaders         = "block_headers"
	tableUTXOs                = "utxos"
	tableXPubs                = "xpubs"
	tablePaymails             = "paymail_addresses"
)

const (
	// ReferenceIDField is used for Paymail
	ReferenceIDField = "reference_id"

	// Internal field names
	broadcastStatusField = "broadcast_status"
	currentBalanceField  = "current_balance"
	draftIDField         = "draft_id"
	idField              = "id"
	metadataField        = "metadata"
	nextExternalNumField = "next_external_num"
	nextInternalNumField = "next_internal_num"
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
	defaultAddressResolutionPurpose = "bux Address Resolution"
	defaultSenderPaymail            = "bitcoinschema@moneybutton.com"
	handleHandcashPrefix            = "$"
	handleMaxLength                 = 25
	handleRelayPrefix               = "1"

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

		// Block Headers as received by the network
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
	}
)
