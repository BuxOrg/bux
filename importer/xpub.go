package importer

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sort"

	"github.com/BuxOrg/bux"
	"github.com/BuxOrg/bux/cachestore"
	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/taskmanager"
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
	gap := 0
	allTransactions := []*whatsonchain.HistoryRecord{}
	// Derive internal addresses until gap limit
	log.Printf("Deriving internal addresses...")
	for i := 0; i < depth; i++ {
		log.Printf("path m/1/%v", i)
		dest, err := buxClient.NewDestination(ctx, xpub.String(), utils.ChainInternal, utils.ScriptTypePubKeyHash, nil)
		if err != nil {
			return err
		}
		// Get history for address
		txs, err := getAddressTransactions(dest.Address)
		if err != nil {
			return err
		}
		if len(txs) == 0 {
			gap++
			continue
		}
		allTransactions = append(allTransactions, txs...)

	}

	// Derive external addresses until gap limit
	log.Printf("Deriving external addresses...")
	for i := 0; i < depth; i++ {
		log.Printf("path m/0/%v", i)
		dest, err := buxClient.NewDestination(ctx, xpub.String(), utils.ChainExternal, utils.ScriptTypePubKeyHash, nil)
		if err != nil {
			return err
		}
		// Get history for address
		txs, err := getAddressTransactions(dest.Address)
		if err != nil {
			return err
		}
		if len(txs) == 0 {
			gap++
			continue
		}
		allTransactions = append(allTransactions, txs...)
	}
	// Remove any duplicate transactions from all historical txs
	allTransactions = removeDuplicates(allTransactions)

	txHashes := whatsonchain.TxHashes{}
	for _, t := range allTransactions {
		txHashes.TxIDs = append(txHashes.TxIDs, t.TxHash)
	}

	// Get all transaction data for each tx in preparation to sort by previous
	// txs
	// TODO: Just get the raw tx and cast it to a bt.Tx. Super inefficient to
	// make more API calls
	/*txInfos := []*whatsonchain.TxInfo{}
	for i := 0; i < len(txHashes.TxIDs); i += 20 {
		num := 20
		bulk := whatsonchain.TxHashes{}
		txSubset := txHashes.TxIDs[i:]
		if len(txSubset) < 20 {
			num = len(txSubset)
		}
		bulk.TxIDs = txSubset[i : i+num-1]

		// Get raw transaction data
		info, err := client.BulkTransactionDetails(context.Background(), &bulk)
		if err != nil {
			log.Printf("ERR: %v", err)
			return err
		}

		txInfos = append(txInfos, info...)
	}*/
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

func recordTransactions(ctx context.Context, rawTxs []string, buxClient bux.ClientInterface) error {
	for _, rawTx := range rawTxs {
		_, err := buxClient.RecordTransaction(ctx, "", rawTx, "")
		if err != nil {
			return err
		}

	}
	return nil
}

func parseXpub(xpubStr string) (*bip32.ExtendedKey, error) {
	return utils.ValidateXPub(xpubStr)
}

func getAddressTransactions(address string) ([]*whatsonchain.HistoryRecord, error) {
	options := whatsonchain.ClientDefaultOptions()
	options.RateLimit = 20
	client := whatsonchain.NewClient(whatsonchain.NetworkMain, options, buildHttpClient())
	history, err := client.AddressHistory(context.TODO(), address)
	if err != nil {
		return nil, err
	}
	txs := []*whatsonchain.HistoryRecord{}
	for _, h := range history {
		txs = append(txs, h)
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

/*func getAddressTransactionsBitbus(address string) ([]string, error) {
	transactions := []string{}
	query := []byte(`
  {
    "q": {
      "find": { "out.e.a": "ADDRESS" },
      "sort": { "blk.i": 1 },
      "project": { "blk": 1, "tx.h": 1, "out.f3": 1 },
      "limit": 100
    }
  }`)
	query = bytes.Replace(query, []byte("ADDRESS"), []byte(address), 1)
	client := http.Client{}
	req, err := http.NewRequest("POST", "https://txo.bitbus.network/block", bytes.NewBuffer(query))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", PlanariaToken)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		json := gjson.ParseBytes(line)
		tx_id := gjson.Get(json.String(), "tx.h")
		transactions = append(transactions, tx_id.Str)
	}
	if err != nil {
		return transactions, err
	}

	return transactions, nil
}*/

func initBuxClient(ctx context.Context, debug bool) (bux.ClientInterface, error) {
	var options []bux.ClientOps
	if debug {
		options = append(options, bux.WithDebugging())
	}
	options = append(options, bux.WithAutoMigrate(bux.BaseModels...))
	options = append(options, bux.WithRistretto(cachestore.DefaultRistrettoConfig()))
	options = append(options, bux.WithTaskQ(taskmanager.DefaultTaskQConfig("imp_queue"), taskmanager.FactoryMemory))
	options = append(options, bux.WithSQLite(&datastore.SQLiteConfig{
		CommonConfig: datastore.CommonConfig{
			Debug:       false,
			TablePrefix: "xapi",
		},
		DatabasePath: "./import.db", // "" for in memory
		Shared:       true,
	}))

	x, err := bux.NewClient(ctx, options...)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return x, err
}
