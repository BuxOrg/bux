package bux

import (
	"context"

	"github.com/mrz1836/go-whatsonchain"
)

// ImportResults are the results from the import
type ImportResults struct {
	ExternalAddresses    int    `json:"external_addresses"`
	InternalAddresses    int    `json:"internal_addresses"`
	Key                  string `json:"key"`
	TransactionsFound    int    `json:"transactions_found"`
	TransactionsImported int    `json:"transactions_imported"`
}

// getTransactionsFromAddresses will get all transactions related to addresses
func getTransactionsFromAddresses(ctx context.Context, client whatsonchain.ClientInterface, addressList whatsonchain.AddressList) ([]*whatsonchain.HistoryRecord, error) {
	histories, err := client.BulkUnspentTransactionsProcessor(
		ctx, &addressList,
	)
	if err != nil {
		return nil, err
	}
	var txs []*whatsonchain.HistoryRecord
	for _, h := range histories {
		txs = append(txs, h.Utxos...)
	}
	return txs, nil
}

// removeDuplicates will remove duplicate transactions
func removeDuplicates(transactions []*whatsonchain.HistoryRecord) []*whatsonchain.HistoryRecord {
	keys := make(map[string]bool)
	var list []*whatsonchain.HistoryRecord

	for _, tx := range transactions {
		if _, value := keys[tx.TxHash]; !value {
			keys[tx.TxHash] = true
			list = append(list, tx)
		}
	}
	return list
}
