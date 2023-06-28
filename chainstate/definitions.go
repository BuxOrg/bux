package chainstate

import (
	"time"

	"github.com/BuxOrg/bux/utils"
)

// Chainstate configuration defaults
const (
	defaultBroadcastTimeOut        = 15 * time.Second
	defaultFalsePositiveRate       = 0.01
	defaultFeeLastCheckIgnore      = 2 * time.Minute
	defaultMaxNumberOfDestinations = 100000
	defaultMonitorDays             = 7
	defaultQueryTimeOut            = 15 * time.Second
	whatsOnChainRateLimitWithKey   = 20
)

const (
	// FilterBloom is for bloom filters
	FilterBloom = "bloom"

	// FilterRegex is for regex filters
	FilterRegex = "regex"
)

// Internal network names
const (
	mainNet    = "mainnet" // Main Public Bitcoin network
	mainNetAlt = "main"    // Main Public Bitcoin network
	stn        = "stn"     // BitcoinSV Public Stress Test Network (https://bitcoinscaling.io/)
	testNet    = "testnet" // Public test network
	testNetAlt = "test"    // Public test network
)

// Requirements and providers
const (
	mAPIFailure       = "failure"  // Minercraft result was a failure / error
	mAPISuccess       = "success"  // Minercraft result was success (still could be an error)
	requiredInMempool = "mempool"  // Requirement for tx query (has to be >= mempool)
	requiredOnChain   = "on-chain" // Requirement for tx query (has to be == on-chain)
)

// List of providers
const (
	ProviderAll          = "all"          // All providers (used for errors etc)
	ProviderMAPI         = "mapi"         // Query & broadcast provider for mAPI (using given miners)
	ProviderNowNodes     = "nownodes"     // Query & broadcast provider for NowNodes
	ProviderWhatsOnChain = "whatsonchain" // Query & broadcast provider for WhatsOnChain
)

// TransactionInfo is the universal information about the transaction found from a chain provider
type TransactionInfo struct {
	BlockHash     string `json:"block_hash,omitempty"`    // mAPI, WOC
	BlockHeight   int64  `json:"block_height"`            // mAPI, WOC
	Confirmations int64  `json:"confirmations,omitempty"` // mAPI, WOC
	ID            string `json:"id"`                      // Transaction ID (Hex)
	MinerID       string `json:"miner_id,omitempty"`      // mAPI ONLY - miner_id found
	Provider      string `json:"provider,omitempty"`      // Provider is our internal source
}

var (
	// DefaultFee is used when a fee has not been set by the user
	// This default is currently accepted by all BitcoinSV miners (50/1000) (7.27.23)
	// Actual TAAL FeeUnit - 1/1000, GorillaPool - 50/1000 (7.27.23)
	DefaultFee = &utils.FeeUnit{
		Satoshis: 1,
		Bytes:    20,
	}
)
