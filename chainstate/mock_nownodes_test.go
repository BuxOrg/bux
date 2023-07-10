package chainstate

import (
	"context"
	"errors"
	"time"

	"github.com/mrz1836/go-nownodes"
)

type nowNodesBase struct{}

func (n *nowNodesBase) GetAddress(context.Context, nownodes.Blockchain, string) (*nownodes.AddressInfo, error) {
	return nil, nil
}

func (n *nowNodesBase) GetTransaction(context.Context, nownodes.Blockchain, string) (*nownodes.TransactionInfo, error) {
	return nil, nil
}

func (n *nowNodesBase) GetMempoolEntry(context.Context, nownodes.Blockchain, string, string) (*nownodes.MempoolEntryResult, error) {
	return nil, nil
}

func (n *nowNodesBase) SendTransaction(context.Context, nownodes.Blockchain, string) (*nownodes.BroadcastResult, error) {
	return nil, nil
}

func (n *nowNodesBase) SendRawTransaction(context.Context, nownodes.Blockchain, string, string) (*nownodes.BroadcastResult, error) {
	return nil, nil
}

func (n *nowNodesBase) HTTPClient() nownodes.HTTPInterface {
	return nil
}

func (n *nowNodesBase) UserAgent() string {
	return ""
}

type nowNodesTxOnChain struct {
	nowNodesBase
}

func (n *nowNodesTxOnChain) SendRawTransaction(context.Context, nownodes.Blockchain,
	string, string,
) (*nownodes.BroadcastResult, error) {
	return nil, errors.New("257: txn-already-known")
}

func (n *nowNodesTxOnChain) GetTransaction(_ context.Context, _ nownodes.Blockchain,
	txID string,
) (*nownodes.TransactionInfo, error) {
	if txID == onChainExample1TxID {
		return &nownodes.TransactionInfo{
			BlockHash:     "0000000000000000015122781ab51d57b26a09518630b882f67f1b08d841979d",
			BlockHeight:   723229,
			BlockTime:     1642777896,
			Confirmations: 314,
			Fees:          "443",
			Hex:           "01000000025b7439a0c9effa3f19d0e441d2eea596e44a8c49240b6e389c29498285f92ad3010000006a4730440220482c1c896678d7307e1de35cef2aae4907f2684617a26d8abd24c444d527c80d02204c550f8f9d69b9cf65780e2e066041750261702639d02605a2eb694ade4ca1d64121029ce7958b2aa3c627334f50bb810c678e2b284db0ef6f7d067f7fccfa05d0f095ffffffff1998b0e4955e1d8ba976d943c43f32e143ba90e805f0e882d3b8edc0f7473b77020000006a47304402204beb486e5d99a15d4d2267e328abb5466a05fdc20d64903d0ace1c4fabb71a34022024803ae9e18b3c11683b2ff2b5fb4ca973a22fdd390f6ab1f99396604a3f06af4121038ea0f258fb838b5193e9739ddd808bb97aaab52a60ba8a83958b13109ab183ccffffffff030000000000000000fd8901006a0372756e0105004d7d017b22696e223a312c22726566223a5b22653864393134303764643461646164363366333739353032303861383532653562306334383037333563656235346133653334333539346163313839616331625f6f31222c22376135346462326162303030306161303035316134383230343162336135653761636239386333363135363863623334393063666564623066653161356438385f6f33225d2c226f7574223a5b2233356463303036313539393333623438353433343565663663633363366261663165666462353263343837313933386632366539313034343632313562343036225d2c2264656c223a5b5d2c22637265223a5b5d2c2265786563223a5b7b226f70223a2243414c4c222c2264617461223a5b7b22246a6967223a307d2c22757064617465222c5b7b22246a6967223a317d2c7b2267726164756174696f6e506f736974696f6e223a6e756c6c2c226c6576656c223a382c226e616d65223a22e38395e383abe38380222c227870223a373030307d5d5d7d5d7d11010000000000001976a914058cae340a2ef8fd2b43a074b75fb6b38cb2765788acd4020000000000001976a914160381a3811b474ff77f31f64f4e57a5bb5ebf1788ac00000000",
			TxID:          onChainExample1TxID,
			Value:         "997",
			ValueIn:       "1440",
			Version:       1,
			Vin: []*nownodes.Input{
				{
					Addresses: []string{"1WLucQHxqVN94QcdVkuARq78dm7JLYk2S"},
					Hex:       "4730440220482c1c896678d7307e1de35cef2aae4907f2684617a26d8abd24c444d527c80d02204c550f8f9d69b9cf65780e2e066041750261702639d02605a2eb694ade4ca1d64121029ce7958b2aa3c627334f50bb810c678e2b284db0ef6f7d067f7fccfa05d0f095",
					IsAddress: true,
					N:         0,
					Sequence:  4294967295,
					TxID:      "d32af9858249299c386e0b24498c4ae496a5eed241e4d0193ffaefc9a039745b",
					Value:     "273",
					VOut:      1,
				},
				{
					Addresses: []string{"13FKuaQuj1aovEfv1smmDZWUGYh39SHT1s"},
					Hex:       "47304402204beb486e5d99a15d4d2267e328abb5466a05fdc20d64903d0ace1c4fabb71a34022024803ae9e18b3c11683b2ff2b5fb4ca973a22fdd390f6ab1f99396604a3f06af4121038ea0f258fb838b5193e9739ddd808bb97aaab52a60ba8a83958b13109ab183cc",
					IsAddress: true,
					N:         1,
					Sequence:  4294967295,
					TxID:      "773b47f7c0edb8d382e8f005e890ba43e1323fc443d976a98b1d5e95e4b09819",
					Value:     "1167",
					VOut:      2,
				},
			},
			VOut: []*nownodes.Output{
				{
					Hex:   "006a0372756e0105004d7d017b22696e223a312c22726566223a5b22653864393134303764643461646164363366333739353032303861383532653562306334383037333563656235346133653334333539346163313839616331625f6f31222c22376135346462326162303030306161303035316134383230343162336135653761636239386333363135363863623334393063666564623066653161356438385f6f33225d2c226f7574223a5b2233356463303036313539393333623438353433343565663663633363366261663165666462353263343837313933386632366539313034343632313562343036225d2c2264656c223a5b5d2c22637265223a5b5d2c2265786563223a5b7b226f70223a2243414c4c222c2264617461223a5b7b22246a6967223a307d2c22757064617465222c5b7b22246a6967223a317d2c7b2267726164756174696f6e506f736974696f6e223a6e756c6c2c226c6576656c223a382c226e616d65223a22e38395e383abe38380222c227870223a373030307d5d5d7d5d7d",
					N:     0,
					Value: "0",
				},
				{
					Addresses: []string{"1WLucQHxqVN94QcdVkuARq78dm7JLYk2S"},
					Hex:       "76a914058cae340a2ef8fd2b43a074b75fb6b38cb2765788ac",
					IsAddress: true,
					N:         1,
					Spent:     true,
					Value:     "273",
				},
				{
					Addresses: []string{"131Q4rG1otRpwHS2CGSo2pDLuN9CkBviDT"},
					Hex:       "76a914160381a3811b474ff77f31f64f4e57a5bb5ebf1788ac",
					IsAddress: true,
					N:         2,
					Spent:     true,
					Value:     "724",
				},
			},
		}, nil
	}

	return nil, nil
}

type nowNodesBroadcastSuccess struct {
	nowNodesBase
}

func (n *nowNodesBroadcastSuccess) SendRawTransaction(_ context.Context, _ nownodes.Blockchain,
	_ string, id string,
) (*nownodes.BroadcastResult, error) {
	return &nownodes.BroadcastResult{
		ID:     id,
		Result: id,
	}, nil
}

type nowNodeInMempool struct {
	nowNodesTxOnChain
}

func (n *nowNodeInMempool) SendRawTransaction(context.Context, nownodes.Blockchain,
	string, string,
) (*nownodes.BroadcastResult, error) {
	return nil, errors.New("257: txn-already-known")
}

type nowNodesTxNotFound struct {
	nowNodesBase
}

func (n *nowNodesTxNotFound) SendRawTransaction(context.Context, nownodes.Blockchain,
	string, string,
) (*nownodes.BroadcastResult, error) {
	return nil, errors.New("56: mempool conflict")
}

func (n *nowNodesTxNotFound) GetTransaction(_ context.Context, _ nownodes.Blockchain,
	txID string,
) (*nownodes.TransactionInfo, error) {
	return nil, errors.New("Transaction '" + txID + "' not found")
}

type nowNodesBroadcastTimeout struct {
	nowNodesBase
}

func (n *nowNodesBroadcastTimeout) SendRawTransaction(context.Context, nownodes.Blockchain,
	string, string,
) (*nownodes.BroadcastResult, error) {
	time.Sleep(defaultBroadcastTimeOut * 2)
	return nil, errors.New("257: txn-already-known")
}
