package chainstate

import (
	"context"
	"errors"

	"github.com/libsv/go-bt/v2"
	"github.com/mrz1836/go-whatsonchain"
)

type whatsOnChainBase struct{}

func (w *whatsOnChainBase) AddressBalance(context.Context, string) (balance *whatsonchain.AddressBalance, err error) {
	return
}

func (w *whatsOnChainBase) AddressHistory(context.Context, string) (history whatsonchain.AddressHistory, err error) {
	return
}

func (w *whatsOnChainBase) AddressInfo(context.Context, string) (addressInfo *whatsonchain.AddressInfo, err error) {
	return
}

func (w *whatsOnChainBase) AddressUnspentTransactionDetails(context.Context, string, int) (history whatsonchain.AddressHistory, err error) {
	return
}

func (w *whatsOnChainBase) AddressUnspentTransactions(context.Context, string) (history whatsonchain.AddressHistory, err error) {
	return
}

func (w *whatsOnChainBase) BroadcastTx(context.Context, string) (txID string, err error) {
	return
}

func (w *whatsOnChainBase) BulkBalance(context.Context, *whatsonchain.AddressList) (balances whatsonchain.AddressBalances, err error) {
	return
}

func (w *whatsOnChainBase) BulkBroadcastTx(context.Context, []string, bool) (response *whatsonchain.BulkBroadcastResponse, err error) {
	return
}

func (w *whatsOnChainBase) BulkScriptUnspentTransactions(context.Context, *whatsonchain.ScriptsList) (response whatsonchain.BulkScriptUnspentResponse, err error) {
	return
}

func (w *whatsOnChainBase) BulkTransactionDetails(context.Context, *whatsonchain.TxHashes) (txList whatsonchain.TxList, err error) {
	return
}

func (w *whatsOnChainBase) BulkTransactionDetailsProcessor(context.Context, *whatsonchain.TxHashes) (txList whatsonchain.TxList, err error) {
	return
}

func (w *whatsOnChainBase) BulkUnspentTransactions(context.Context, *whatsonchain.AddressList) (response whatsonchain.BulkUnspentResponse, err error) {
	return
}

func (w *whatsOnChainBase) BulkUnspentTransactionsProcessor(context.Context, *whatsonchain.AddressList) (response whatsonchain.BulkUnspentResponse, err error) {
	return
}

func (w *whatsOnChainBase) DecodeTransaction(context.Context, string) (txInfo *whatsonchain.TxInfo, err error) {
	return
}

func (w *whatsOnChainBase) DownloadReceipt(context.Context, string) (string, error) {
	return "", nil
}

func (w *whatsOnChainBase) DownloadStatement(context.Context, string) (string, error) {
	return "", nil
}

func (w *whatsOnChainBase) GetBlockByHash(context.Context, string) (blockInfo *whatsonchain.BlockInfo, err error) {
	return
}

func (w *whatsOnChainBase) GetBlockByHeight(context.Context, int64) (blockInfo *whatsonchain.BlockInfo, err error) {
	return
}

func (w *whatsOnChainBase) GetBlockPages(context.Context, string, int) (txList whatsonchain.BlockPagesInfo, err error) {
	return
}

func (w *whatsOnChainBase) GetChainInfo(context.Context) (chainInfo *whatsonchain.ChainInfo, err error) {
	return
}

func (w *whatsOnChainBase) GetCirculatingSupply(context.Context) (supply float64, err error) {
	return
}

func (w *whatsOnChainBase) GetExchangeRate(context.Context) (rate *whatsonchain.ExchangeRate, err error) {
	return
}

func (w *whatsOnChainBase) GetExplorerLinks(context.Context, string) (results whatsonchain.SearchResults, err error) {
	return
}

func (w *whatsOnChainBase) GetHeaderByHash(context.Context, string) (headerInfo *whatsonchain.BlockInfo, err error) {
	return
}

func (w *whatsOnChainBase) GetHeaders(context.Context) (blockHeaders []*whatsonchain.BlockInfo, err error) {
	return
}

func (w *whatsOnChainBase) GetHealth(context.Context) (string, error) {
	return "", nil
}

func (w *whatsOnChainBase) GetMempoolInfo(context.Context) (info *whatsonchain.MempoolInfo, err error) {
	return
}

func (w *whatsOnChainBase) GetMempoolTransactions(context.Context) (transactions []string, err error) {
	return
}

func (w *whatsOnChainBase) GetMerkleProof(context.Context, string) (merkleResults whatsonchain.MerkleResults, err error) {
	return
}

func (w *whatsOnChainBase) GetMerkleProofTSC(context.Context, string) (merkleResults whatsonchain.MerkleTSCResults, err error) {
	return
}

func (w *whatsOnChainBase) GetRawTransactionData(context.Context, string) (string, error) {
	return "", nil
}

func (w *whatsOnChainBase) BulkRawTransactionDataProcessor(context.Context, *whatsonchain.TxHashes) (whatsonchain.TxList, error) {
	return nil, nil
}

func (w *whatsOnChainBase) GetRawTransactionOutputData(context.Context, string, int) (string, error) {
	return "", nil
}

func (w *whatsOnChainBase) GetScriptHistory(context.Context, string) (history whatsonchain.ScriptList, err error) {
	return
}

func (w *whatsOnChainBase) GetScriptUnspentTransactions(context.Context, string) (scriptList whatsonchain.ScriptList, err error) {
	return
}

func (w *whatsOnChainBase) GetTxByHash(context.Context, string) (txInfo *whatsonchain.TxInfo, err error) {
	return
}

func (w *whatsOnChainBase) HTTPClient() whatsonchain.HTTPInterface {
	return nil
}

func (w *whatsOnChainBase) LastRequest() *whatsonchain.LastRequest {
	return nil
}

func (w *whatsOnChainBase) Network() whatsonchain.NetworkType {
	return whatsonchain.NetworkMain
}

func (w *whatsOnChainBase) UserAgent() string {
	return "default-user-agent"
}

func (w *whatsOnChainBase) RateLimit() int {
	return 3
}

type whatsOnChainTxOnChain struct {
	whatsOnChainBase
}

func (w *whatsOnChainTxOnChain) BroadcastTx(context.Context, string) (string, error) {
	return "", errors.New("unexpected response code 500: 257: txn-already-known")
}

func (w *whatsOnChainTxOnChain) GetTxByHash(_ context.Context, hash string) (txInfo *whatsonchain.TxInfo, err error) {

	if hash == onChainExample1TxID {
		txInfo = &whatsonchain.TxInfo{
			BlockHash:     onChainExample1BlockHash,
			BlockHeight:   onChainExample1BlockHeight,
			BlockTime:     1642777896,
			Confirmations: onChainExample1Confirmations,
			Hash:          onChainExample1TxID,
			Hex: "01000000025b7439a0c9effa3f19d0e441d2eea596e44a8c49240b6e389c29498285f92ad3010000006a4730440" +
				"220482c1c896678d7307e1de35cef2aae4907f2684617a26d8abd24c444d527c80d02204c550f8f9d69b9cf65780e2e0660417" +
				"50261702639d02605a2eb694ade4ca1d64121029ce7958b2aa3c627334f50bb810c678e2b284db0ef6f7d067f7fccfa05d0f09" +
				"5ffffffff1998b0e4955e1d8ba976d943c43f32e143ba90e805f0e882d3b8edc0f7473b77020000006a47304402204beb486e5" +
				"d99a15d4d2267e328abb5466a05fdc20d64903d0ace1c4fabb71a34022024803ae9e18b3c11683b2ff2b5fb4ca973a22fdd390" +
				"f6ab1f99396604a3f06af4121038ea0f258fb838b5193e9739ddd808bb97aaab52a60ba8a83958b13109ab183ccffffffff030" +
				"000000000000000fd8901006a0372756e0105004d7d017b22696e223a312c22726566223a5b226538643931343037646434616" +
				"461643633663337393530323038613835326535623063343830373335636562353461336533343335393461633138396163316" +
				"25f6f31222c2237613534646232616230303030616130303531613438323034316233613565376163623938633336313536386" +
				"3623334393063666564623066653161356438385f6f33225d2c226f7574223a5b2233356463303036313539393333623438353" +
				"433343565663663633363366261663165666462353263343837313933386632366539313034343632313562343036225d2c226" +
				"4656c223a5b5d2c22637265223a5b5d2c2265786563223a5b7b226f70223a2243414c4c222c2264617461223a5b7b22246a696" +
				"7223a307d2c22757064617465222c5b7b22246a6967223a317d2c7b2267726164756174696f6e506f736974696f6e223a6e756" +
				"c6c2c226c6576656c223a382c226e616d65223a22e38395e383abe38380222c227870223a373030307d5d5d7d5d7d110100000" +
				"00000001976a914058cae340a2ef8fd2b43a074b75fb6b38cb2765788acd4020000000000001976a914160381a3811b474ff77" +
				"f31f64f4e57a5bb5ebf1788ac00000000",
			LockTime: 0,
			Size:     776,
			Time:     1642777896,
			TxID:     onChainExample1TxID,
			Version:  1,
			// NOTE: no vin / vout since they are not being used
		}
	}
	return
}

type whatsOnChainBroadcastSuccess struct {
	whatsOnChainBase
}

func (w *whatsOnChainBroadcastSuccess) BroadcastTx(_ context.Context, hex string) (string, error) {
	tx, err := bt.NewTxFromString(hex)
	if err != nil {
		return "", err
	}

	return tx.TxID(), nil
}

type whatsOnChainInMempool struct {
	whatsOnChainBase
}

func (w *whatsOnChainInMempool) BroadcastTx(_ context.Context, hex string) (string, error) {
	return "", errors.New("unexpected response code 500: 257: txn-already-known")
}

type whatsOnChainTxNotFound struct {
	whatsOnChainBase
}

func (w *whatsOnChainTxNotFound) GetTxByHash(context.Context, string) (txInfo *whatsonchain.TxInfo, err error) {
	return nil, whatsonchain.ErrTransactionNotFound
}

func (w *whatsOnChainTxNotFound) BroadcastTx(context.Context, string) (string, error) {
	return "", errors.New("unexpected response code 500: mempool conflict")
}
