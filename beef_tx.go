package bux

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
)

func ToBeefHex(tx *Transaction) (string, error) {
	beef, err := newBeefTx(1, tx)
	if err != nil {
		return "", err
	}

	beefBytes, err := beef.toBeefBytes()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(beefBytes), nil
}

type beefTx struct {
	version             uint32
	marker              []byte
	compoundMerklePaths CMPSlice
	transactions        []*Transaction
}

func newBeefTx(version uint32, tx *Transaction) (*beefTx, error) {
	// get inputs previous transactions
	inputs := tx.draftTransaction.Configuration.Inputs
	transactions := make([]*Transaction, 0, len(inputs)+1)

	for _, input := range inputs {
		prevTx, err := getInputPrevTransaction(tx.client, input)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, prevTx)
	}

	// add current transaction
	transactions = append(transactions, tx)

	beef := &beefTx{
		version:             version,
		marker:              []byte{0x00, 0x00, 0x00, 0x00, 0xBE, 0xEF},
		compoundMerklePaths: tx.draftTransaction.CompoundMerklePathes,
		transactions:        khanTopologicalSort(transactions),
	}

	return beef, nil
}

func getInputPrevTransaction(client ClientInterface, input *TransactionInput) (*Transaction, error) {
	inputTx, err := client.GetTransactionByID(context.Background(), input.UtxoPointer.TransactionID)
	if err != nil {
		return nil, err
	}
	if inputTx.MerkleProof.TxOrID != "" {
		return inputTx, nil
	} else {
		return nil, errors.New("transaction is not mined yet") // TODO: handle it in next iterration
	}
}

func (tx *beefTx) toBeefBytes() ([]byte, error) {
	if tx.compoundMerklePaths == nil || tx.transactions == nil {
		return nil, errors.New("beef tx is incomplete")
	}

	// get beef bytes
	beefSize := 0

	version := bt.LittleEndianBytes(tx.version, 4)
	beefSize += len(version)

	beefMarker := tx.marker
	beefSize += len(beefMarker)

	nPaths := bt.VarInt(len(tx.compoundMerklePaths)).Bytes()
	beefSize += len(nPaths)

	compoundMerklePaths := make([][]byte, 0, len(tx.compoundMerklePaths))

	for i, cmp := range tx.compoundMerklePaths {
		compoundMerklePaths = append(compoundMerklePaths, cmp.Bytes())
		beefSize += len(compoundMerklePaths[i])
	}

	nTransactions := bt.VarInt(uint64(len(tx.transactions))).Bytes()
	beefSize += len(nTransactions)

	transactions := make([][]byte, 0, len(tx.transactions))

	for i, t := range tx.transactions {
		transactions = append(transactions, t.toBeefBytes(tx.compoundMerklePaths))
		beefSize += len(transactions[i])
	}

	// compose beef
	buffer := make([]byte, 0, beefSize)
	buffer = append(buffer, version...)
	buffer = append(buffer, beefMarker...)
	buffer = append(buffer, nPaths...)

	for _, cmp := range compoundMerklePaths {
		buffer = append(buffer, cmp...)
	}

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
	buffer := make([]byte, beefSize)
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

func btInputToCefBytes(in *bt.Input, compountedPaths CMPSlice) []byte {
	h := make([]byte, 0)

	// Raw Tx Input data
	h = append(h, in.Bytes(false)...)

	// get PathIndex
	pathIdx := -1
	previousTxId := binary.LittleEndian.Uint64(in.PreviousTxID())
	for i, cmp := range compountedPaths {
		for _, path := range cmp {
			for _, txId := range path {
				if txId == previousTxId {
					pathIdx = i
				}
			}
		}
	}

	if pathIdx > -1 {
		// CEF byte marker
		h = append(h, byte(0xEF)) // it's extended if has pathIndex

		// Path Index
		h = append(h, bt.VarInt(pathIdx).Bytes()...)

		// Previous input satoshis
		h = append(h, utils.LittleEndianBytes64(in.PreviousTxSatoshis, 8)...)

		if in.PreviousTxScript != nil {
			prevTxScriptLen := uint64(len(*in.PreviousTxScript))
			h = append(h, bt.VarInt(prevTxScriptLen).Bytes()...)
			h = append(h, *in.PreviousTxScript...)
		} else {
			h = append(h, 0x00) // The length of the script is zero
		}
	} else {
		// No extended data
		// CEF byte marker
		h = append(h, byte(0x00))
	}

	return h
}

func khanTopologicalSort(transactions []*Transaction) []*Transaction {
	// create a map to store transactions by their IDs for random access.
	transactionMap := make(map[string]*Transaction)

	// init the in-degree and result slices.
	inDegree := make(map[string]int)
	queue := make([]string, 0)
	result := make([]*Transaction, 0)

	// init in-degree map and transaction map.
	for _, tx := range transactions {
		transactionMap[tx.ID] = tx
		inDegree[tx.ID] = 0
	}

	// calculate in-degrees.
	for _, tx := range transactions {
		for _, input := range tx.draftTransaction.Configuration.Inputs {
			inDegree[input.UtxoPointer.TransactionID]++
		}
	}

	// enqueue transactions with in-degree of 0.
	for txID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, txID)
		}
	}

	// topological sort
	for len(queue) > 0 {
		txID := queue[0]
		queue = queue[1:]
		result = append(result, transactionMap[txID])

		// update in-degrees and enqueue neighbors.
		for _, input := range transactionMap[txID].draftTransaction.Configuration.Inputs {
			neighborId := input.UtxoPointer.TransactionID
			inDegree[neighborId]--
			if inDegree[neighborId] == 0 {
				queue = append(queue, neighborId)
			}
		}
	}

	// reverse sorted collection
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}
