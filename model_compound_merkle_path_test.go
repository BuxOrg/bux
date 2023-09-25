package bux

import (
	"testing"

	"github.com/libsv/go-bt/v2"
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
			[]map[string]bt.VarInt{
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
			[]map[string]bt.VarInt{
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
		assert.Equal(t, CompoundMerklePath{}, cmp)
	})
}

// TestCompoundMerklePathModel_Hex will test the method Hex()
func TestCompoundMerklePathModel_Hex(t *testing.T) {
	t.Run("Sorted Compound Merkle Path", func(t *testing.T) {
		cmp := CompoundMerklePath(
			[]map[string]bt.VarInt{
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
		expectedHex := "030200ABCDEFGH01IJKLMNOP0300ABCD01EFGH02IJKL0300AB02EF07OP0602txId103D06G07txId20cM0dtxId3"
		actualHex := cmp.Hex()
		assert.Equal(t, expectedHex, actualHex)
	})

	t.Run("Unsorted Compound Merkle Path", func(t *testing.T) {
		cmp := CompoundMerklePath(
			[]map[string]bt.VarInt{
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
		expectedHex := "020200ABCD01EFGH0400AB01CD02EF03GH0800A01B02C03D04E05F06G07H"
		actualHex := cmp.Hex()
		assert.Equal(t, expectedHex, actualHex)
	})
}

func TestCompoundMerklePathModel_CalculateCompoundMerklePathAndCalculateHex(t *testing.T) {
	t.Parallel()

	t.Run("Real Merkle Proof", func(t *testing.T) {
		signleMerkleProof := []MerkleProof{
			{
				Index:  1153,
				TxOrID: "2130b63dcbfe1356a30137fe9578691f59c6cf42d5e8928a800619de7f8e14da",
				Nodes: []string{
					"4d4bde1dc35c87bba992944ec0379e0bb009916108113dc3de1c4aecda6457a3",
					"168595f83accfcec66d0e0df06df89e6a9a2eaa3aa69427fb86cb54d8ea5b1e9",
					"c2edd41b237844a45a0e6248a9e7c520af303a5c91cc8a443ad0075d6a3dec79",
					"bdd0fddf45fee49324e55dfc6fdb9044c86dc5be3dbf941a80b395838495ac09",
					"3e5ec052b86621b5691d15ad54fab2551c27a36d9ab84f428a304b607aa33d33",
					"9feb9b1aaa2cd8486edcacb60b9d477a89aec5867d292608c3c59a18324d608a",
					"22e1db219f8d874315845b7cee84832dc0865b5f9e18221a011043a4d6704e7d",
					"7f118890abd8df3f8a51c344da0f9235609f5fd380e38cfe519e81262aedb2a7",
					"20dcf60bbcecd2f587e8d3344fb68c71f2f2f7a6cc85589b9031c2312a433fe6",
					"0be65c1f3b53b937608f8426e43cb41c1db31227d0d9933e8b0ce3b8cc30d67f",
					"a8036cf77d8de296f60607862b228174733a30486a37962a56465f5e8c214d87",
					"b8e4d7975537bb775e320f01f874c06cf38dd2ce7bb836a1afe0337aeb9fb06f",
					"88e6b0bd93e02b057ea43a80a5bb8cf9673f143340af3f569fe0c55c085e5efb",
					"15f731176e17f4402802d5be3893419e690225e732d69dfd27f6e614f188233d",
				},
			},
		}
		expectedCMP := CompoundMerklePath(
			[]map[string]bt.VarInt{
				{
					"4d4bde1dc35c87bba992944ec0379e0bb009916108113dc3de1c4aecda6457a3": 1152,
					"2130b63dcbfe1356a30137fe9578691f59c6cf42d5e8928a800619de7f8e14da": 1153,
				},
				{
					"168595f83accfcec66d0e0df06df89e6a9a2eaa3aa69427fb86cb54d8ea5b1e9": 577,
				},
				{
					"c2edd41b237844a45a0e6248a9e7c520af303a5c91cc8a443ad0075d6a3dec79": 289,
				},
				{
					"bdd0fddf45fee49324e55dfc6fdb9044c86dc5be3dbf941a80b395838495ac09": 145,
				},
				{
					"3e5ec052b86621b5691d15ad54fab2551c27a36d9ab84f428a304b607aa33d33": 73,
				},
				{
					"9feb9b1aaa2cd8486edcacb60b9d477a89aec5867d292608c3c59a18324d608a": 37,
				},
				{
					"22e1db219f8d874315845b7cee84832dc0865b5f9e18221a011043a4d6704e7d": 19,
				},
				{
					"7f118890abd8df3f8a51c344da0f9235609f5fd380e38cfe519e81262aedb2a7": 8,
				},
				{
					"20dcf60bbcecd2f587e8d3344fb68c71f2f2f7a6cc85589b9031c2312a433fe6": 5,
				},
				{
					"0be65c1f3b53b937608f8426e43cb41c1db31227d0d9933e8b0ce3b8cc30d67f": 3,
				},
				{
					"a8036cf77d8de296f60607862b228174733a30486a37962a56465f5e8c214d87": 0,
				},
				{
					"b8e4d7975537bb775e320f01f874c06cf38dd2ce7bb836a1afe0337aeb9fb06f": 1,
				},
				{
					"88e6b0bd93e02b057ea43a80a5bb8cf9673f143340af3f569fe0c55c085e5efb": 1,
				},
				{
					"15f731176e17f4402802d5be3893419e690225e732d69dfd27f6e614f188233d": 1,
				},
			},
		)
		cmp, err := CalculateCompoundMerklePath(signleMerkleProof)
		assert.NoError(t, err)
		assert.Equal(t, expectedCMP, cmp)
		expectedHex := "0d" + //13 - height
			"01" + // nLeafs at this height VarInt
			"01" + // offset VarInt
			"15f731176e17f4402802d5be3893419e690225e732d69dfd27f6e614f188233d" + // 32 byte hash
			// ----------------------
			// implied end of leaves at this height
			// height of next leaves is therefore 12
			"01" + // nLeafs at this height VarInt
			"01" + // offset VarInt
			"88e6b0bd93e02b057ea43a80a5bb8cf9673f143340af3f569fe0c55c085e5efb" + // 32 byte hash
			// ----------------------
			// implied end of leaves at this height
			// height of next leaves is therefore 11 and so on...
			"01" +
			"01" +
			"b8e4d7975537bb775e320f01f874c06cf38dd2ce7bb836a1afe0337aeb9fb06f" +
			"01" +
			"00" +
			"a8036cf77d8de296f60607862b228174733a30486a37962a56465f5e8c214d87" +
			"01" +
			"03" +
			"0be65c1f3b53b937608f8426e43cb41c1db31227d0d9933e8b0ce3b8cc30d67f" +
			"01" +
			"05" +
			"20dcf60bbcecd2f587e8d3344fb68c71f2f2f7a6cc85589b9031c2312a433fe6" +
			"01" +
			"08" +
			"7f118890abd8df3f8a51c344da0f9235609f5fd380e38cfe519e81262aedb2a7" +
			"01" +
			"13" +
			"22e1db219f8d874315845b7cee84832dc0865b5f9e18221a011043a4d6704e7d" +
			"01" +
			"25" +
			"9feb9b1aaa2cd8486edcacb60b9d477a89aec5867d292608c3c59a18324d608a" +
			"01" +
			"49" +
			"3e5ec052b86621b5691d15ad54fab2551c27a36d9ab84f428a304b607aa33d33" +
			"01" +
			"91" +
			"bdd0fddf45fee49324e55dfc6fdb9044c86dc5be3dbf941a80b395838495ac09" +
			"01" +
			"fd2101" +
			"c2edd41b237844a45a0e6248a9e7c520af303a5c91cc8a443ad0075d6a3dec79" +
			"01" +
			"fd4102" +
			"168595f83accfcec66d0e0df06df89e6a9a2eaa3aa69427fb86cb54d8ea5b1e9" +
			"02" +
			"fd8004" +
			"4d4bde1dc35c87bba992944ec0379e0bb009916108113dc3de1c4aecda6457a3" +
			"fd8104" +
			"2130b63dcbfe1356a30137fe9578691f59c6cf42d5e8928a800619de7f8e14da"

		actualHex := cmp.Hex()
		assert.Equal(t, expectedHex, actualHex)
	})
}
