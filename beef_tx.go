package bux

import (
	"context"
	"encoding/hex"
	"errors"
)

// ToBeefHex generates BEEF Hex for transaction
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
	compoundMerklePaths CMPSlice
	transactions        []*Transaction
}

func newBeefTx(version uint32, tx *Transaction) (*beefTx, error) {
	// get inputs parent transactions
	inputs := tx.draftTransaction.Configuration.Inputs
	transactions := make([]*Transaction, 0, len(inputs)+1)

	for _, input := range inputs {
		prevTx, err := getParentTransactionForInput(tx.client, input)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, prevTx)
	}

	// add current transaction
	transactions = append(transactions, tx)

	beef := &beefTx{
		version:             version,
		compoundMerklePaths: tx.draftTransaction.CompoundMerklePathes,
		transactions:        kahnTopologicalSortTransactions(transactions),
	}

	return beef, nil
}

func getParentTransactionForInput(client ClientInterface, input *TransactionInput) (*Transaction, error) {
	inputTx, err := client.GetTransactionByID(context.Background(), input.UtxoPointer.TransactionID)
	if err != nil {
		return nil, err
	}

	if inputTx.MerkleProof.TxOrID != "" {
		return inputTx, nil
	}

	return nil, errors.New("transaction is not mined yet") // TODO: handle it in next iterration
}
