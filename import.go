package bux

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/BuxOrg/bux/utils"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/tidwall/gjson"
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
// getUnspentTransactionsFromAddresses will get all unspent transactions related to address
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

/*
// getAllTransactionsFromAddresses will get all transactions related to addresses
func getAllTransactionsFromAddresses(ctx context.Context, client whatsonchain.ClientInterface,
	addressList whatsonchain.AddressList) ([]*whatsonchain.HistoryRecord, []string, error) {
	var addressesWithTransactions []string
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
*/

// deriveAddresses will derive a new set of addresses for a xpub
func (c *Client) deriveAddresses(ctx context.Context, xpub string, chain uint32, amount int) ([]string, error) {
	var addressList []string
	for i := 0; i < amount; i++ {
		destination, err := c.NewDestination(ctx, xpub, chain, utils.ScriptTypePubKeyHash, false, c.DefaultModelOptions()...)
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

func getTransactionsFromAddressesViaBitbus(addresses []string) ([]*whatsonchain.HistoryRecord, error) {
	var transactions []*whatsonchain.HistoryRecord
	parentQuery := []byte(`
  {
    "q": {
			"find": { "$or": [ { "out.e.a": "ADDRESS" }, {"in.e.a": "ADDRESS"} ] },
      "sort": { "blk.i": 1 },
      "project": { "blk": 1, "tx.h": 1, "out.f3": 1 },
      "limit": LIMIT,
			"skip": OFFSET
    }
  }`)
	for _, address := range addresses {
		doneWithPagination := false
		offset := 0
		limit := 100
		for !doneWithPagination {
			query := bytes.Replace(parentQuery, []byte("ADDRESS"), []byte(address), 2)
			query = bytes.Replace(query, []byte("OFFSET"), []byte(strconv.Itoa(offset)), 1)
			query = bytes.Replace(query, []byte("LIMIT"), []byte(strconv.Itoa(limit)), 1)
			txs, err := bitbusRequest(query)
			if err != nil {
				return transactions, err
			}
			transactions = append(transactions, txs...)
			if len(txs) < 100 {
				doneWithPagination = true
			}
			offset += 100
		}
	}

	return transactions, nil
}

func bitbusRequest(query []byte) ([]*whatsonchain.HistoryRecord, error) {
	planariaToken := os.Getenv("PLANARIA_TOKEN")
	client := http.Client{}
	var transactions []*whatsonchain.HistoryRecord
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, "https://txo.bitbus.network/block", bytes.NewBuffer(query),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", planariaToken)

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	reader := bufio.NewReader(resp.Body)
	for {
		var line []byte
		line, err = reader.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			break
		}
		json := gjson.ParseBytes(line)
		txID := gjson.Get(json.String(), "tx.h")
		if txID.Str == "" {
			break
		}
		record := &whatsonchain.HistoryRecord{
			TxHash: txID.Str,
		}
		transactions = append(transactions, record)
	}
	return transactions, nil
}
