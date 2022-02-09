package chainstate

import (
	"context"
	"errors"

	"github.com/libsv/go-bt/v2"
	"github.com/mrz1836/go-mattercloud"
)

type matterCloudBase struct{}

func (mt *matterCloudBase) Network() mattercloud.NetworkType {
	return mattercloud.NetworkMain
}

func (mt *matterCloudBase) AddressBalance(context.Context, string) (balance *mattercloud.Balance, err error) {
	return
}

func (mt *matterCloudBase) AddressBalanceBatch(context.Context, []string) (balances []*mattercloud.Balance, err error) {
	return
}

func (mt *matterCloudBase) AddressHistory(context.Context, string) (history *mattercloud.History, err error) {
	return
}

func (mt *matterCloudBase) AddressHistoryBatch(context.Context, []string) (history *mattercloud.History, err error) {
	return
}

func (mt *matterCloudBase) AddressUtxos(context.Context, string) (utxos []*mattercloud.UnspentTransaction, err error) {
	return
}

func (mt *matterCloudBase) AddressUtxosBatch(context.Context, []string) (utxos []*mattercloud.UnspentTransaction, err error) {
	return
}

func (mt *matterCloudBase) Broadcast(context.Context, string) (response *mattercloud.BroadcastResponse, err error) {
	return
}

func (mt *matterCloudBase) Transaction(context.Context, string) (transaction *mattercloud.Transaction, err error) {
	return
}

func (mt *matterCloudBase) TransactionBatch(context.Context, []string) (transactions []*mattercloud.Transaction, err error) {
	return
}

func (mt *matterCloudBase) Request(context.Context, string, string, []byte) (response string, err error) {
	return
}

type matterCloudTxOnChain struct {
	matterCloudBase
}

func (mt *matterCloudTxOnChain) Broadcast(context.Context, string) (*mattercloud.BroadcastResponse, error) {
	/*
		{
		    "success": false,
		    "code": 500,
		    "error": "GENERAL_ERROR: ERROR: Missing inputs",
		    "message": "GENERAL_ERROR: ERROR: Missing inputs"
		}
	*/
	return nil, errors.New("GENERAL_ERROR: ERROR: Missing inputs")
}

func (mt *matterCloudTxOnChain) Transaction(_ context.Context,
	tx string) (transaction *mattercloud.Transaction, err error) {

	if tx == onChainExample1TxID {
		transaction = &mattercloud.Transaction{
			BlockHash:     onChainExample1BlockHash,
			BlockHeight:   onChainExample1BlockHeight,
			BlockTime:     1642777896,
			Confirmations: onChainExample1Confirmations,
			Fees:          0.00000443,
			Hash:          onChainExample1TxID,
			LockTime:      0,
			RawTx: "01000000025b7439a0c9effa3f19d0e441d2eea596e44a8c49240b6e389c29498285f92ad3010000006a4730440" +
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
			Size:     776,
			Time:     1642777896,
			TxID:     onChainExample1TxID,
			ValueIn:  0.0000144,
			ValueOut: 0.00000997,
			Version:  1,
			// NOTE: no vin / vout since they are not being used
		}
	}

	return
}

type matterCloudBroadcastSuccess struct {
	matterCloudBase
}

func (mt *matterCloudBroadcastSuccess) Broadcast(_ context.Context, hex string) (*mattercloud.BroadcastResponse, error) {
	tx, err := bt.NewTxFromString(hex)
	if err != nil {
		return nil, err
	}

	return &mattercloud.BroadcastResponse{
		Success: true,
		Result:  &mattercloud.BroadcastResult{TxID: tx.TxID()},
	}, nil
}

type matterCloudInMempool struct {
	matterCloudBase
}

func (mt *matterCloudInMempool) Broadcast(_ context.Context, hex string) (*mattercloud.BroadcastResponse, error) {
	/*{
	    "success": false,
	    "code": 422,
	    "error": "TXN-ALREADY-KNOWN",
	    "message": "TXN-ALREADY-KNOWN"
	}*/
	return nil, errors.New("TXN-ALREADY-KNOWN")
}

type matterCloudTxNotFound struct {
	matterCloudBase
}

func (mt *matterCloudTxNotFound) Transaction(context.Context,
	string) (transaction *mattercloud.Transaction, err error) {
	return nil, errors.New("unexpected error occurred")
}

func (mt *matterCloudTxNotFound) Broadcast(context.Context, string) (*mattercloud.BroadcastResponse, error) {
	/*
		{
		    "success": false,
		    "code": 500,
		    "error": "GENERAL_ERROR: ERROR: Missing inputs",
		    "message": "GENERAL_ERROR: ERROR: Missing inputs"
		}
	*/
	return nil, errors.New("GENERAL_ERROR: ERROR: Mempool conflict")
}
