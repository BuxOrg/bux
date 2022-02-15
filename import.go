package bux

import (
	"context"

	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-whatsonchain"
)

// ImportResults are the results from the import
type ImportResults struct {
	ExternalAddresses         int      `json:"external_addresses"`
	InternalAddresses         int      `json:"internal_addresses"`
	AddressesWithTransactions []string `json:"addresses_with_transactions"`
	Key                       string   `json:"key"`
	TransactionsFound         int      `json:"transactions_found"`
	TransactionsImported      int      `json:"transactions_imported"`
}

/*
// getUnspentTransactionsFromAddresses will get all unspent transactions related to addresses
func getUnspentTransactionsFromAddresses(ctx context.Context, client whatsonchain.ClientInterface, addressList whatsonchain.AddressList) ([]*whatsonchain.HistoryRecord, error) {
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
*/

// getAllTransactionsFromAddresses will get all transactions related to addresses
func getAllTransactionsFromAddresses(ctx context.Context, client whatsonchain.ClientInterface, addressList whatsonchain.AddressList) ([]*whatsonchain.HistoryRecord, []string, error) {
	addressesWithTransactions := []string{}
	var txs []*whatsonchain.HistoryRecord
	for _, address := range addressList.Addresses {
		history, err := client.AddressHistory(ctx, address)
		if err != nil {
			return nil, addressesWithTransactions, err
		}
		if len(history) > 0 {
			addressesWithTransactions = append(addressesWithTransactions, address)
		}
		txs = append(txs, history...)
	}
	return txs, addressesWithTransactions, nil
}

// deriveAddresses will derive a new set of addresses for an xpub
func (c *Client) deriveAddresses(ctx context.Context, xpub string, chain uint32, amount int) ([]string, error) {
	addressList := []string{}
	for i := 0; i < amount; i++ {
		destination, err := c.NewDestination(ctx, xpub, chain, utils.ScriptTypePubKeyHash, nil)
		if err != nil {
			return []string{}, err
		}
		addressList = append(addressList, destination.Address)
	}
	return addressList, nil
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
