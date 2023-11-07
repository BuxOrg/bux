package bux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			Path: [][]BUMPLeaf{
				{
					{Offset: 0, Hash: "node0"},
					{Offset: 1, Hash: "txId", TxID: true},
				},
				{
					{Offset: 1, Hash: "node1"},
				},
				{
					{Offset: 1, Hash: "node2"},
				},
				{
					{Offset: 1, Hash: "node3"},
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
			Path: [][]BUMPLeaf{
				{
					{Offset: 14, Hash: "txId", TxID: true},
					{Offset: 15, Hash: "node0"},
				},
				{
					{Offset: 6, Hash: "node1"},
				},
				{
					{Offset: 2, Hash: "node2"},
				},
				{
					{Offset: 0, Hash: "node3"},
				},
				{
					{Offset: 1, Hash: "node4"},
				},
			},
		}
		actualBUMP := mp.ToBUMP()
		assert.Equal(t, expectedBUMP, actualBUMP)
	})

	t.Run("Valid Merkle Proof #3 - with *", func(t *testing.T) {
		mp := MerkleProof{
			Index:  14,
			TxOrID: "txId",
			Nodes:  []string{"*", "node1", "node2", "node3", "node4"},
		}
		expectedBUMP := BUMP{
			Path: [][]BUMPLeaf{
				{
					{Offset: 14, Hash: "txId", TxID: true},
					{Offset: 15, Duplicate: true},
				},
				{
					{Offset: 6, Hash: "node1"},
				},
				{
					{Offset: 2, Hash: "node2"},
				},
				{
					{Offset: 0, Hash: "node3"},
				},
				{
					{Offset: 1, Hash: "node4"},
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
