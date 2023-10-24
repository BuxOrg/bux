package bux

import (
	"testing"

	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
)

// TestMerkleProofModel_ToCompoundMerklePath will test the method ToCompoundMerklePath()
func TestMerkleProofModel_ToCompoundMerklePath(t *testing.T) {
	t.Parallel()

	t.Run("Valid Merkle Proof #1", func(t *testing.T) {
		mp := MerkleProof{
			Index:  1,
			TxOrID: "txId",
			Nodes:  []string{"node0", "node1", "node2", "node3"},
		}
		expectedCMP := CompoundMerklePath(
			[]map[string]bt.VarInt{
				{
					"node0": 0,
					"txId":  1,
				},
				{
					"node1": 1,
				},
				{
					"node2": 1,
				},
				{
					"node3": 1,
				},
			},
		)
		cmp := mp.ToCompoundMerklePath()
		assert.Equal(t, expectedCMP, cmp)
	})

	t.Run("Valid Merkle Proof #2", func(t *testing.T) {
		mp := MerkleProof{
			Index:  14,
			TxOrID: "txId",
			Nodes:  []string{"node0", "node1", "node2", "node3", "node4"},
		}
		expectedCMP := CompoundMerklePath(
			[]map[string]bt.VarInt{
				{
					"txId":  14,
					"node0": 15,
				},
				{
					"node1": 6,
				},
				{
					"node2": 2,
				},
				{
					"node3": 0,
				},
				{
					"node4": 1,
				},
			},
		)
		cmp := mp.ToCompoundMerklePath()
		assert.Equal(t, expectedCMP, cmp)
	})

	t.Run("Empty Merkle Proof", func(t *testing.T) {
		mp := MerkleProof{}
		cmp := mp.ToCompoundMerklePath()
		assert.Nil(t, cmp)
	})
}

// TestMerkleProofModel_ToBUMP will test the method ToBUMP()
func TestMerkleProofModel_ToBUMP(t *testing.T) {
	t.Parallel()

	t.Run("Valid Merkle Proof #1", func(t *testing.T) {
		mp := MerkleProof{
			Index:  1,
			TxOrID: "txId",
			Nodes:  []string{"node0", "node1", "node2", "node3"},
		}
		expectedBUMP := BUMP{
			Path: []BUMPPathMap{
				{
					"0": BUMPPathElement{Hash: "node0"},
					"1": BUMPPathElement{Hash: "txId", TxId: true},
				},
				{
					"1": BUMPPathElement{Hash: "node1"},
				},
				{
					"1": BUMPPathElement{Hash: "node2"},
				},
				{
					"1": BUMPPathElement{Hash: "node3"},
				},
			},
		}
		actualBUMP := mp.ToBUMP()
		assert.Equal(t, expectedBUMP, actualBUMP)
	})

	t.Run("Valid Merkle Proof #2", func(t *testing.T) {
		mp := MerkleProof{
			Index:  14,
			TxOrID: "txId",
			Nodes:  []string{"node0", "node1", "node2", "node3", "node4"},
		}
		expectedBUMP := BUMP{
			Path: []BUMPPathMap{
				{
					"14": BUMPPathElement{Hash: "txId", TxId: true},
					"15": BUMPPathElement{Hash: "node0"},
				},
				{
					"6": BUMPPathElement{Hash: "node1"},
				},
				{
					"2": BUMPPathElement{Hash: "node2"},
				},
				{
					"0": BUMPPathElement{Hash: "node3"},
				},
				{
					"1": BUMPPathElement{Hash: "node4"},
				},
			},
		}
		actualBUMP := mp.ToBUMP()
		assert.Equal(t, expectedBUMP, actualBUMP)
	})

	t.Run("Empty Merkle Proof", func(t *testing.T) {
		mp := MerkleProof{}
		actualBUMP := mp.ToBUMP()
		assert.Equal(t, BUMP{}, actualBUMP)
	})
}
