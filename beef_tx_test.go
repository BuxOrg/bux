package bux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// func Test_toBeefBytes(t *testing.T) {
// 	// input prev transaction
// 	inputs := make([]*TransactionInput, 0)
// 	for _, inTxId := range inputsTxIds {
// 		in := &TransactionInput{
// 			Utxo: Utxo{
// 				UtxoPointer: UtxoPointer{
// 					TransactionID: inTxId,
// 				},
// 			},
// 		}

// 		inputs = append(inputs, in)
// 	}

// 	inputPrevTransaction:= *Transaction{

// 	}

// 	beefTransaction := &beefTx{
// 		version:             1,
// 		marker:              []byte{0x00, 0x00, 0x00, 0x00, 0xBE, 0xEF}, // beef marker from BRC62
// 		compoundMerklePaths: nil,
// 		transactions:        nil,
// 	}

// 	expectedHex := "0100000000000000BEEF01020101cd73c0c6bb645581816fa960fd2f1636062fcbf23cb57981074ab8d708a76e3b02003470d882cf556a4b943639eba15dc795dffdbebdc98b9a98e3637fda96e3811e01c58e40f22b9e9fcd05a09689a9b19e6e62dbfd3335c5253d09a7a7cd755d9a3c0201da256f78ae0ad74bbf539662cdb9122aa02ba9a9d883f1d52468d96290515adb02b4c8d919190a090e77b73ffcd52b85babaaeeb62da000473102aca7f070facef02020000000158cb8b052fded9a6c450c4212562df8820359ec34da41286421e0d0f2b7eefee000000006a47304402206b1255cb23454c63b22833de25a3a3ecbdb8d8645ad129d3269cdddf10b2ec98022034cadf46e5bfecc38940e5497ddf5fa9aeb37ff5ec3fe8e21b19cbb64a45ec324121029a82bfce319faccc34095c8405896e1223921917501a4f736a04f126d6a01c12ffffffffef00c6c7c903000000001976a9145e506dd7afa889e8f16f5e00555d1c0ab225152f88ac0101000000000000001976a914d866ec5ebb0f4e3840351ee61887101e5407562988ac00000000020000000158cb8b052fded9a6c450c4212562df8820359ec34da41286421e0d0f2b7eefee000000006a47304402206b1255cb23454c63b22833de25a3a3ecbdb8d8645ad129d3269cdddf10b2ec98022034cadf46e5bfecc38940e5497ddf5fa9aeb37ff5ec3fe8e21b19cbb64a45ec324121029a82bfce319faccc34095c8405896e1223921917501a4f736a04f126d6a01c12ffffffff000101000000000000001976a914d866ec5ebb0f4e3840351ee61887101e5407562988ac00000000"

// 	t.Run("beef bytes test", func(t *testing.T) {

// 		beefBytes, err := beefTransaction.toBeefBytes()

// 		assert.Nil(t, err)
// 		assert.Equal(t, expectedHex, hex.EncodeToString(beefBytes))
// 	})

// }

func Test_khanTopologicalSort(t *testing.T) {
	// create related transactions from oldest to newest
	txsFromOldestToNewest := []*Transaction{
		createTx("0"),
		createTx("1", "0"),
		createTx("2", "1"),
		createTx("3", "2", "1"),
		createTx("4", "3", "1"),
	}

	unsortedTxs := []*Transaction{
		txsFromOldestToNewest[2],
		txsFromOldestToNewest[3],
		txsFromOldestToNewest[0],
		txsFromOldestToNewest[4],
		txsFromOldestToNewest[1],
	}

	t.Run("khanTopologicalSort sort from oldest to newest", func(t *testing.T) {
		sortedGraph := khanTopologicalSort(unsortedTxs)

		for i, tx := range sortedGraph {
			assert.Equal(t, tx.ID, sortedGraph[i].ID)
		}

	})
}

func createTx(txId string, inputsTxIds ...string) *Transaction {
	inputs := make([]*TransactionInput, 0)
	for _, inTxId := range inputsTxIds {
		in := &TransactionInput{
			Utxo: Utxo{
				UtxoPointer: UtxoPointer{
					TransactionID: inTxId,
				},
			},
		}

		inputs = append(inputs, in)
	}

	transaction := &Transaction{
		draftTransaction: &DraftTransaction{
			Configuration: TransactionConfig{
				Inputs: inputs,
			},
		},
	}

	transaction.ID = txId

	return transaction
}
