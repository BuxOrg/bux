package bux

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/libsv/go-bt/v2"
)

var hasCmp = byte(0x01)
var hasNoCmp = byte(0x00)

func (beefTx *beefTx) toBeefBytes() ([]byte, error) {
	if len(beefTx.compoundMerklePaths) == 0 || len(beefTx.transactions) < 2 { // valid BEEF contains atleast two transactions (new transaction and one parent transaction)
		return nil, errors.New("beef tx is incomplete")
	}

	// get beef bytes
	beefSize := 0

	version := bt.LittleEndianBytes(beefTx.version, 4)
	version[2] = 0xBE
	version[3] = 0xEF
	beefSize += len(version)

	nPaths := bt.VarInt(len(beefTx.compoundMerklePaths)).Bytes()
	beefSize += len(nPaths)

	compoundMerklePaths := beefTx.compoundMerklePaths.Bytes()
	beefSize += len(compoundMerklePaths)

	nTransactions := bt.VarInt(uint64(len(beefTx.transactions))).Bytes()
	beefSize += len(nTransactions)

	transactions := make([][]byte, 0, len(beefTx.transactions))

	for _, t := range beefTx.transactions {
		txBytes, err := t.toBeefBytes(beefTx.compoundMerklePaths)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, txBytes)
		beefSize += len(txBytes)
	}

	// compose beef
	buffer := make([]byte, 0, beefSize)
	buffer = append(buffer, version...)
	buffer = append(buffer, nPaths...)
	buffer = append(buffer, compoundMerklePaths...)

	buffer = append(buffer, nTransactions...)

	for _, t := range transactions {
		buffer = append(buffer, t...)
	}

	return buffer, nil
}

func (tx *Transaction) toBeefBytes(compountedPaths CMPSlice) ([]byte, error) {
	txBeefBytes, err := hex.DecodeString(tx.Hex)

	if err != nil {
		return nil, fmt.Errorf("decoding tx (ID: %s) hex failed: %w", tx.ID, err)
	}

	cmpIdx := tx.getCompountedMarklePathIndex(compountedPaths)
	if cmpIdx > -1 {
		txBeefBytes = append(txBeefBytes, hasCmp)
		txBeefBytes = append(txBeefBytes, bt.VarInt(cmpIdx).Bytes()...)
	} else {
		txBeefBytes = append(txBeefBytes, hasNoCmp)
	}

	return txBeefBytes, nil
}

func (tx *Transaction) getCompountedMarklePathIndex(compountedPaths CMPSlice) int {
	pathIdx := -1

	for i, cmp := range compountedPaths {
		for txID := range cmp[0] {
			if txID == tx.ID {
				pathIdx = i
			}
		}
	}

	return pathIdx
}
