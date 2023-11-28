package chainstate

// Validate validates TransactionInfo by checking if it contains
// BlockHash and MerkleProof (from mAPI) or MerklePath (from Arc)
func (t *TransactionInfo) Validate() bool {
	if t.BlockHash == "" ||
		((t.MerkleProof == nil || t.MerkleProof.TxOrID == "" || len(t.MerkleProof.Nodes) == 0) &&
			(t.MerklePath == nil || len(t.MerklePath.Path) == 0)) {
		return false
	}
	return true
}
