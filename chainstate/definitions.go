package chainstate

import (
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bc"
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
	ProviderAll             = "all"             // All providers (used for errors etc)
	ProviderMAPI            = "mapi"            // Query & broadcast provider for mAPI (using given miners)
	ProviderNowNodes        = "nownodes"        // Query & broadcast provider for NowNodes
	ProviderWhatsOnChain    = "whatsonchain"    // Query & broadcast provider for WhatsOnChain
	ProviderBroadcastClient = "broadcastclient" // Query & broadcast provider for configured miners
)

// TransactionInfo is the universal information about the transaction found from a chain provider
type TransactionInfo struct {
	BlockHash     string          `json:"block_hash,omitempty"`    // mAPI, WOC
	BlockHeight   int64           `json:"block_height"`            // mAPI, WOC
	Confirmations int64           `json:"confirmations,omitempty"` // mAPI, WOC
	ID            string          `json:"id"`                      // Transaction ID (Hex)
	MinerID       string          `json:"miner_id,omitempty"`      // mAPI ONLY - miner_id found
	Provider      string          `json:"provider,omitempty"`      // Provider is our internal source
	MerkleProof   *bc.MerkleProof `json:"merkle_proof,omitempty"`  // mAPI 1.5 ONLY. Should be also supported by Arc in future
}

// DefaultFee is used when a fee has not been set by the user
// This default is currently accepted by all BitcoinSV miners (50/1000) (7.27.23)
// Actual TAAL FeeUnit - 1/1000, GorillaPool - 50/1000 (7.27.23)
var DefaultFee = &utils.FeeUnit{
	Satoshis: 1,
	Bytes:    20,
}

// BlockInfo is the response info about a returned block
type BlockInfo struct {
	Bits              string         `json:"bits"`
	ChainWork         string         `json:"chainwork"`
	CoinbaseTx        CoinbaseTxInfo `json:"coinbaseTx"`
	Confirmations     int64          `json:"confirmations"`
	Difficulty        float64        `json:"difficulty"`
	Hash              string         `json:"hash"`
	Height            int64          `json:"height"`
	MedianTime        int64          `json:"mediantime"`
	MerkleRoot        string         `json:"merkleroot"`
	Miner             string         `json:"Bmgpool"`
	NextBlockHash     string         `json:"nextblockhash"`
	Nonce             int64          `json:"nonce"`
	Pages             Page           `json:"pages"`
	PreviousBlockHash string         `json:"previousblockhash"`
	Size              int64          `json:"size"`
	Time              int64          `json:"time"`
	TotalFees         float64        `json:"totalFees"`
	Tx                []string       `json:"tx"`
	TxCount           int64          `json:"txcount"`
	Version           int64          `json:"version"`
	VersionHex        string         `json:"versionHex"`
}

// CoinbaseTxInfo is the coinbase tx info inside the BlockInfo
type CoinbaseTxInfo struct {
	BlockHash     string     `json:"blockhash"`
	BlockTime     int64      `json:"blocktime"`
	Confirmations int64      `json:"confirmations"`
	Hash          string     `json:"hash"`
	Hex           string     `json:"hex"`
	LockTime      int64      `json:"locktime"`
	Size          int64      `json:"size"`
	Time          int64      `json:"time"`
	TxID          string     `json:"txid"`
	Version       int64      `json:"version"`
	Vin           []VinInfo  `json:"vin"`
	Vout          []VoutInfo `json:"vout"`
}

// Page is used as a subtype for BlockInfo
type Page struct {
	Size int64    `json:"size"`
	URI  []string `json:"uri"`
}

// VinInfo is the vin info inside the CoinbaseTxInfo
type VinInfo struct {
	Coinbase  string        `json:"coinbase"`
	ScriptSig ScriptSigInfo `json:"scriptSig"`
	Sequence  int64         `json:"sequence"`
	TxID      string        `json:"txid"`
	Vout      int64         `json:"vout"`
}

// VoutInfo is the vout info inside the CoinbaseTxInfo
type VoutInfo struct {
	N            int64            `json:"n"`
	ScriptPubKey ScriptPubKeyInfo `json:"scriptPubKey"`
	Value        float64          `json:"value"`
}

// ScriptSigInfo is the scriptSig info inside the VinInfo
type ScriptSigInfo struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

// ScriptPubKeyInfo is the scriptPubKey info inside the VoutInfo
type ScriptPubKeyInfo struct {
	Addresses   []string `json:"addresses"`
	Asm         string   `json:"asm"`
	Hex         string   `json:"hex"`
	IsTruncated bool     `json:"isTruncated"`
	OpReturn    string   `json:"-"` // todo: support this (can be an object of key/vals based on the op return data)
	ReqSigs     int64    `json:"reqSigs"`
	Type        string   `json:"type"`
}

// TxInfo is the response info about a returned tx
type TxInfo struct {
	BlockHash     string     `json:"blockhash"`
	BlockHeight   int64      `json:"blockheight"`
	BlockTime     int64      `json:"blocktime"`
	Confirmations int64      `json:"confirmations"`
	Hash          string     `json:"hash"`
	Hex           string     `json:"hex"`
	LockTime      int64      `json:"locktime"`
	Size          int64      `json:"size"`
	Time          int64      `json:"time"`
	TxID          string     `json:"txid"`
	Version       int64      `json:"version"`
	Vin           []VinInfo  `json:"vin"`
	Vout          []VoutInfo `json:"vout"`

	Error string `json:"error"`
}
