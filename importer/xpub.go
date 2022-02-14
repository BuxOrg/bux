package importer

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"sort"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt"
	"github.com/mrz1836/go-whatsonchain"
)

var WhatsOnChainApiKey = os.Getenv("WOC_API_KEY")

func ImportXpub(ctx context.Context, buxClient bux.ClientInterface, xpub *bip32.ExtendedKey, depth, gapLimit int, path string) error {
	options := whatsonchain.ClientDefaultOptions()
	options.RateLimit = 20
	client := whatsonchain.NewClient(whatsonchain.NetworkMain, options, buildHttpClient())
	allTransactions := []*whatsonchain.HistoryRecord{}

	addressList := whatsonchain.AddressList{}
	// Derive internal addresses until gap limit
	log.Printf("Deriving internal addresses...")
	for i := 0; i < depth; i++ {
		log.Printf("path m/1/%v", i)
		dest, err := buxClient.NewDestination(ctx, xpub.String(), utils.ChainInternal, utils.ScriptTypePubKeyHash, nil)
		if err != nil {
			return err
		}
		addressList.Addresses = append(addressList.Addresses, dest.Address)
	}

	// Derive external addresses until gap limit
	log.Printf("Deriving external addresses...")
	for i := 0; i < depth; i++ {
		log.Printf("path m/0/%v", i)
		dest, err := buxClient.NewDestination(ctx, xpub.String(), utils.ChainExternal, utils.ScriptTypePubKeyHash, nil)
		if err != nil {
			return err
		}
		addressList.Addresses = append(addressList.Addresses, dest.Address)
	}

	allTransactions, err := getAddressesTransactions(addressList)
	if err != nil {
		return err
	}

	// Remove any duplicate transactions from all historical txs
	allTransactions = removeDuplicates(allTransactions)

	txHashes := whatsonchain.TxHashes{}
	for _, t := range allTransactions {
		txHashes.TxIDs = append(txHashes.TxIDs, t.TxHash)
	}

	rawTxs := []string{}
	txInfos, err := client.BulkRawTransactionDataProcessor(context.Background(), &txHashes)
	if err != nil {
		return err
	}
	for i := 0; i < len(txInfos); i++ {
		tx, err := bt.NewTxFromString(txInfos[i].Hex)
		if err != nil {
			return err
		}
		vins := []whatsonchain.VinInfo{}
		for _, in := range tx.Inputs {
			vin := whatsonchain.VinInfo{
				TxID: in.PreviousTxID,
			}
			vins = append(vins, vin)
		}
		txInfos[i].Vin = vins
		rawTxs = append(rawTxs, txInfos[i].Hex)
	}
	log.Printf("Sorting transactions to be recorded...")
	// Sort all transactions by block height
	sort.Slice(txInfos, func(i, j int) bool {
		return txInfos[i].BlockHeight < txInfos[j].BlockHeight
	})

	// Sort transactions that are in the same block by previous tx
	for i := 0; i < len(txInfos); i++ {
		info := txInfos[i]
		bh := info.BlockHeight
		sameBlockTxs := []*whatsonchain.TxInfo{}
		sameBlockTxs = append(sameBlockTxs, info)
		// Loop through all remaining txs until block height is not the same
		for j := i + 1; j < len(txInfos); j++ {
			if txInfos[j].BlockHeight == bh {
				sameBlockTxs = append(sameBlockTxs, txInfos[j])
			} else {
				break
			}
		}
		if len(sameBlockTxs) == 1 {
			continue
		}
		// Sort transactions by whether or not previous txs are referenced in the inputs
		sort.Slice(sameBlockTxs, func(i, j int) bool {
			for _, in := range sameBlockTxs[i].Vin {
				if in.TxID == sameBlockTxs[j].Hash {
					return false
				}
			}
			return true
		})
		copy(txInfos[i:i+len(sameBlockTxs)], sameBlockTxs)
		i += len(sameBlockTxs) - 1
	}
	// Record transactions in bux
	err = recordTransactions(ctx, rawTxs, buxClient)
	if err != nil {
		log.Printf("ERR: %v", err)
	}
	return nil
}

// Remove duplicate transactions
func removeDuplicates(transactions []*whatsonchain.HistoryRecord) []*whatsonchain.HistoryRecord {
	keys := make(map[string]bool)
	list := []*whatsonchain.HistoryRecord{}

	for _, tx := range transactions {
		if _, value := keys[tx.TxHash]; !value {
			keys[tx.TxHash] = true
			list = append(list, tx)
		}
	}
	return list
}

// Record transactions into database
func recordTransactions(ctx context.Context, rawTxs []string, buxClient bux.ClientInterface) error {
	for _, rawTx := range rawTxs {
		_, err := buxClient.RecordTransaction(ctx, "", rawTx, "")
		if err != nil {
			return err
		}

	}
	return nil
}

// Get all transactions related to an address
func getAddressesTransactions(addressList whatsonchain.AddressList) ([]*whatsonchain.HistoryRecord, error) {
	options := whatsonchain.ClientDefaultOptions()
	options.RateLimit = 20
	client := whatsonchain.NewClient(whatsonchain.NetworkMain, options, buildHttpClient())
	histories, err := client.BulkUnspentTransactionsProcessor(context.TODO(), &addressList)
	if err != nil {
		return nil, err
	}
	txs := []*whatsonchain.HistoryRecord{}
	for _, h := range histories {
		txs = append(txs, h.Utxos...)
	}
	return txs, nil
}

func buildHttpClient() *http.Client {
	options := whatsonchain.ClientDefaultOptions()
	// dial is the net dialer for clientDefaultTransport
	dial := &net.Dialer{KeepAlive: options.DialerKeepAlive, Timeout: options.DialerTimeout}

	// clientDefaultTransport is the default transport struct for the HTTP client
	clientDefaultTransport := &http.Transport{
		DialContext:           dial.DialContext,
		ExpectContinueTimeout: options.TransportExpectContinueTimeout,
		IdleConnTimeout:       options.TransportIdleTimeout,
		MaxIdleConns:          options.TransportMaxIdleConnections,
		Proxy:                 http.ProxyFromEnvironment,
		TLSHandshakeTimeout:   options.TransportTLSHandshakeTimeout,
	}
	tr := &customTransport{apiKey: WhatsOnChainApiKey, rt: clientDefaultTransport}

	return &http.Client{Transport: tr}
}

type customTransport struct {
	apiKey string
	// keep a reference to the client's original transport
	rt http.RoundTripper
}

func (t *customTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// set your auth headers here
	r.Header.Set("woc-api-key", t.apiKey)
	return t.rt.RoundTrip(r)
}
