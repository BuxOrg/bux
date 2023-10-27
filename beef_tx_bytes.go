package bux

import (
	"errors"

	"github.com/libsv/go-bt/v2"
)

var (
	hasBUMP   = byte(0x01)
	hasNoBUMP = byte(0x00)
)

func (beefTx *beefTx) toBeefBytes() ([]byte, error) {
	if len(beefTx.bumpPaths) == 0 || len(beefTx.transactions) < 2 { // valid BEEF contains at least two transactions (new transaction and one parent transaction)
		return nil, errors.New("beef tx is incomplete")
	}

	// get beef bytes
	beefSize := 0

	ver := bt.LittleEndianBytes(beefTx.version, 4)
	ver[2] = 0xBE
	ver[3] = 0xEF
	beefSize += len(ver)

	nBUMPS := bt.VarInt(len(beefTx.bumpPaths)).Bytes()
	beefSize += len(nBUMPS)

	bumps := beefTx.bumpPaths.Bytes()
	beefSize += len(bumps)

	nTransactions := bt.VarInt(uint64(len(beefTx.transactions))).Bytes()
	beefSize += len(nTransactions)

	transactions := make([][]byte, 0, len(beefTx.transactions))

	for _, t := range beefTx.transactions {
		txBytes, err := toBeefBytes(t, beefTx.bumpPaths)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, txBytes)
		beefSize += len(txBytes)
	}

	// compose beef
	buffer := make([]byte, 0, beefSize)
	buffer = append(buffer, version...)
	buffer = append(buffer, nBUMPS...)
	buffer = append(buffer, bumps...)

	buffer = append(buffer, nTransactions...)

	for _, t := range transactions {
		buffer = append(buffer, t...)
	}

	return buffer, nil
}

func toBeefBytes(tx *bt.Tx, bumps BUMPPaths) ([]byte, error) {
	txBeefBytes := tx.Bytes()

	cmpIdx := getBumpPathIndex(tx, bumps)
	if cmpIdx > -1 {
		txBeefBytes = append(txBeefBytes, hasBUMP)
		txBeefBytes = append(txBeefBytes, bt.VarInt(cmpIdx).Bytes()...)
	} else {
		txBeefBytes = append(txBeefBytes, hasNoBUMP)
	}

	return txBeefBytes, nil
}

func getBumpPathIndex(tx *bt.Tx, bumps BUMPPaths) int {
	bumpIndex := -1

	for i, bump := range bumps {
		for txID := range bump.Path[0] {
			if txID == txID {
				bumpIndex = i
			}
		}
	}

	return bumpIndex
}
