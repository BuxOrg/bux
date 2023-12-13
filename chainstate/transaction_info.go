package chainstate

import (
	"github.com/libsv/go-bc"
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

// Valid validates TransactionInfo by checking if it contains
// BlockHash and MerkleProof (from mAPI) or MerklePath (from Arc)
func (t *TransactionInfo) Valid() bool {
	return !(t.BlockHash == "" || t.MerkleProof == nil || t.MerkleProof.TxOrID == "" || len(t.MerkleProof.Nodes) == 0)
}
