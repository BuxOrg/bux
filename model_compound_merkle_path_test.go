package bux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCompoundMerklePathModel_CalculateCompoundMerklePath will test the method CalculateCompoundMerklePath()
func TestCompoundMerklePathModel_CalculateCompoundMerklePath(t *testing.T) {
	t.Parallel()

	t.Run("Single Merkle Proof", func(t *testing.T) {
		signleMerkleProof := []MerkleProof{
			{
				Index:  1,
				TxOrID: "txId",
				Nodes:  []string{"node0", "node1", "node2", "node3"},
			},
		}
		expectedCMP := CompoundMerklePath(
			[]map[string]uint64{
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
		cmp, err := CalculateCompoundMerklePath(signleMerkleProof)
		assert.NoError(t, err)
		assert.Equal(t, expectedCMP, cmp)
	})

	t.Run("Slice of Merkle Proofs", func(t *testing.T) {
		signleMerkleProof := []MerkleProof{
			{
				Index:  2,
				TxOrID: "txId1",
				Nodes:  []string{"D", "AB", "EFGH", "IJKLMNOP"},
			},
			{
				Index:  7,
				TxOrID: "txId2",
				Nodes:  []string{"G", "EF", "ABCD", "IJKLMNOP"},
			},
			{
				Index:  13,
				TxOrID: "txId3",
				Nodes:  []string{"M", "OP", "IJKL", "ABCDEFGH"},
			},
		}
		expectedCMP := CompoundMerklePath(
			[]map[string]uint64{
				{
					"txId1": 2,
					"D":     3,
					"G":     6,
					"txId2": 7,
					"M":     12,
					"txId3": 13,
				},
				{
					"AB": 0,
					"EF": 2,
					"OP": 7,
				},
				{
					"ABCD": 0,
					"EFGH": 1,
					"IJKL": 2,
				},
				{
					"ABCDEFGH": 0,
					"IJKLMNOP": 1,
				},
			},
		)
		cmp, err := CalculateCompoundMerklePath(signleMerkleProof)
		assert.NoError(t, err)
		assert.Equal(t, expectedCMP, cmp)
	})

	t.Run("Paired Transactions", func(t *testing.T) {
		signleMerkleProof := []MerkleProof{
			{
				Index:  8,
				TxOrID: "I",
				Nodes:  []string{"J", "KL", "MNOP", "ABCDEFGH"},
			},
			{
				Index:  9,
				TxOrID: "J",
				Nodes:  []string{"I", "KL", "MNOP", "ABCDEFGH"},
			},
		}
		expectedCMP := CompoundMerklePath(
			[]map[string]uint64{
				{
					"I": 8,
					"J": 9,
				},
				{
					"KL": 5,
				},
				{
					"MNOP": 3,
				},
				{
					"ABCDEFGH": 0,
				},
			},
		)
		cmp, err := CalculateCompoundMerklePath(signleMerkleProof)
		assert.NoError(t, err)
		assert.Equal(t, expectedCMP, cmp)
	})

	t.Run("Different sizes of Merkle Proofs", func(t *testing.T) {
		signleMerkleProof := []MerkleProof{
			{
				Index:  8,
				TxOrID: "I",
				Nodes:  []string{"J", "KL", "MNOP", "ABCDEFGH"},
			},
			{
				Index:  9,
				TxOrID: "J",
				Nodes:  []string{"I", "KL", "MNOP"},
			},
		}
		cmp, err := CalculateCompoundMerklePath(signleMerkleProof)
		assert.Error(t, err)
		assert.Nil(t, cmp)
	})

	t.Run("Empty slice of Merkle Proofs", func(t *testing.T) {
		signleMerkleProof := []MerkleProof{}
		cmp, err := CalculateCompoundMerklePath(signleMerkleProof)
		assert.NoError(t, err)
		assert.Equal(t, cmp, CompoundMerklePath{})
	})

	t.Run("Slice of empty Merkle Proofs", func(t *testing.T) {
		signleMerkleProof := []MerkleProof{
			{}, {}, {},
		}
		cmp, err := CalculateCompoundMerklePath(signleMerkleProof)
		assert.NoError(t, err)
		assert.Equal(t, cmp, CompoundMerklePath{})
	})
}

// TestCompoundMerklePathModel_Hex will test the method Hex()
func TestCompoundMerklePathModel_Hex(t *testing.T) {
	t.Run("Sorted Compound Merkle Path", func(t *testing.T) {
		cmp := CompoundMerklePath(
			[]map[string]uint64{
				{
					"txId1": 2,
					"D":     3,
					"G":     6,
					"txId2": 7,
					"M":     12,
					"txId3": 13,
				},
				{
					"AB": 0,
					"EF": 2,
					"OP": 7,
				},
				{
					"ABCD": 0,
					"EFGH": 1,
					"IJKL": 2,
				},
				{
					"ABCDEFGH": 0,
					"IJKLMNOP": 1,
				},
			},
		)
		expectedHex := "040602txId103D06G07txId212M13txId30300AB02EF07OP0300ABCD01EFGH02IJKL0200ABCDEFGH01IJKLMNOP"
		hex := cmp.Hex()
		assert.Equal(t, hex, expectedHex)
	})

	t.Run("Unsorted Compound Merkle Path", func(t *testing.T) {
		cmp := CompoundMerklePath(
			[]map[string]uint64{
				{
					"F": 5,
					"E": 4,
					"C": 2,
					"D": 3,
					"G": 6,
					"H": 7,
					"B": 1,
					"A": 0,
				},
				{
					"GH": 3,
					"AB": 0,
					"EF": 2,
					"CD": 1,
				},
				{
					"ABCD": 0,
					"EFGH": 1,
				},
			},
		)
		expectedHex := "030800A01B02C03D04E05F06G07H0400AB01CD02EF03GH0200ABCD01EFGH"
		hex := cmp.Hex()
		assert.Equal(t, hex, expectedHex)
	})
}
