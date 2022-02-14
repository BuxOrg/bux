package chainstate

import "time"

// Chainstate configuration defaults
const (
	defaultBroadcastTimeOut      = 15 * time.Second
	defaultQueryTimeOut          = 15 * time.Second
	defaultUserAgent             = "go-chainstate: " + version
	version                      = "v0.1.0"
	whatsOnChainRateLimitWithKey = 20
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
	mAPIFailure          = "failure"      // Minercraft result was a failure / error
	mAPISuccess          = "success"      // Minercraft result was success (still could be an error)
	providerMatterCloud  = "mattercloud"  // Query & broadcast provider for MatterCloud
	providerNowNodes     = "nownodes"     // Query & broadcast provider for NowNodes
	providerWhatsOnChain = "whatsonchain" // Query & broadcast provider for WhatsOnChain
	requiredInMempool    = "mempool"      // Requirement for tx query (has to be >= mempool)
	requiredOnChain      = "on-chain"     // Requirement for tx query (has to be == on-chain)
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
