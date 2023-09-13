package bux

import (
	"encoding/binary"
	"errors"

	"github.com/libsv/go-bt/v2"
)

var beefMarker = []byte{0x00, 0x00, 0x00, 0x00, 0xBE, 0xEF}

func (beefTx *beefTx) toBeefBytes() ([]byte, error) {
	if beefTx.compoundMerklePaths == nil || beefTx.transactions == nil {
		return nil, errors.New("beef tx is incomplete")
	}

	// get beef bytes
	beefSize := 0

	version := bt.LittleEndianBytes(beefTx.version, 4)
	beefSize += len(version)
	beefSize += len(beefMarker)

	nPaths := bt.VarInt(len(beefTx.compoundMerklePaths)).Bytes()
	beefSize += len(nPaths)

	compoundMerklePaths := beefTx.compoundMerklePaths.Bytes()
	beefSize += len(compoundMerklePaths)

	nTransactions := bt.VarInt(uint64(len(beefTx.transactions))).Bytes()
	beefSize += len(nTransactions)

	transactions := make([][]byte, 0, len(beefTx.transactions))

	for i, t := range beefTx.transactions {
		transactions = append(transactions, t.toBeefBytes(beefTx.compoundMerklePaths))
		beefSize += len(transactions[i])
	}

	// compose beef
	buffer := make([]byte, 0, beefSize)
	buffer = append(buffer, version...)
	buffer = append(buffer, beefMarker...)
	buffer = append(buffer, nPaths...)
	buffer = append(buffer, compoundMerklePaths...)

	buffer = append(buffer, nTransactions...)

	for _, t := range transactions {
		buffer = append(buffer, t...)
	}

	return buffer, nil
}

func (tx *Transaction) toBeefBytes(compountedPaths CMPSlice) []byte {
	bttx, _ := bt.NewTxFromString(tx.Hex)

	// get beef bytes
	beefSize := 0

	version := bt.LittleEndianBytes(bttx.Version, 4)
	beefSize += len(version)

	inCounter := bt.VarInt(uint64(len(bttx.Inputs))).Bytes()
	beefSize += len(inCounter)

	inputs := make([][]byte, 0, len(bttx.Inputs))
	for i, in := range bttx.Inputs {
		inputs = append(inputs, btInputToCefBytes(in, compountedPaths))
		beefSize += len(inputs[i])
	}

	outCounter := bt.VarInt(uint64(len(bttx.Outputs))).Bytes()
	beefSize += len(outCounter)

	outputs := make([][]byte, 0, len(bttx.Outputs))
	for i, out := range bttx.Outputs {
		outputs = append(outputs, out.Bytes())
		beefSize += len(outputs[i])
	}

	nLock := make([]byte, 4)
	binary.LittleEndian.PutUint32(nLock, bttx.LockTime)
	beefSize += len(nLock)

	// compose beef
	buffer := make([]byte, 0, beefSize)
	buffer = append(buffer, version...)
	buffer = append(buffer, inCounter...)

	for _, in := range inputs {
		buffer = append(buffer, in...)
	}

	buffer = append(buffer, outCounter...)

	for _, out := range outputs {
		buffer = append(buffer, out...)
	}

	buffer = append(buffer, nLock...)

	return buffer
}
