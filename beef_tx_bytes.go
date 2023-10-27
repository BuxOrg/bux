package bux

import (
	"errors"

	"github.com/libsv/go-bt/v2"
)

var hasCmp = byte(0x01)
var hasNoCmp = byte(0x00)

func (beefTx *beefTx) toBeefBytes() ([]byte, error) {
	if len(beefTx.compoundMerklePaths) == 0 || len(beefTx.transactions) < 2 { // valid BEEF contains at least two transactions (new transaction and one parent transaction)
		return nil, errors.New("beef tx is incomplete")
	}

	// get beef bytes
	beefSize := 0

	ver := bt.LittleEndianBytes(beefTx.version, 4)
	ver[2] = 0xBE
	ver[3] = 0xEF
	beefSize += len(ver)

	nPaths := bt.VarInt(len(beefTx.compoundMerklePaths)).Bytes()
	beefSize += len(nPaths)

	compoundMerklePaths := beefTx.compoundMerklePaths.Bytes()
	beefSize += len(compoundMerklePaths)

	nTransactions := bt.VarInt(uint64(len(beefTx.transactions))).Bytes()
	beefSize += len(nTransactions)

	transactions := make([][]byte, 0, len(beefTx.transactions))

	for _, t := range beefTx.transactions {
		txBytes := toBeefBytes(t, beefTx.compoundMerklePaths)

		transactions = append(transactions, txBytes)
		beefSize += len(txBytes)
	}

	// compose beef
	buffer := make([]byte, 0, beefSize)
	buffer = append(buffer, ver...)
	buffer = append(buffer, nPaths...)
	buffer = append(buffer, compoundMerklePaths...)

	buffer = append(buffer, nTransactions...)

	for _, t := range transactions {
		buffer = append(buffer, t...)
	}

	return buffer, nil
}

func toBeefBytes(tx *bt.Tx, compountedPaths CMPSlice) []byte {
	txBeefBytes := tx.Bytes()

	cmpIdx := getCompountedMarklePathIndex(tx, compountedPaths)
	if cmpIdx > -1 {
		txBeefBytes = append(txBeefBytes, hasCmp)
		txBeefBytes = append(txBeefBytes, bt.VarInt(cmpIdx).Bytes()...)
	} else {
		txBeefBytes = append(txBeefBytes, hasNoCmp)
	}

	return txBeefBytes
}

func getCompountedMarklePathIndex(tx *bt.Tx, compountedPaths CMPSlice) int {
	pathIdx := -1

	for i, cmp := range compountedPaths {
		for txID := range cmp[0] {
			if txID == tx.TxID() {
				pathIdx = i
			}
		}
	}

	return pathIdx
}
