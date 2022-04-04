package chainstate

import (
	"github.com/BuxOrg/bux/utils"
	"github.com/centrifugal/centrifuge-go"
	"github.com/stretchr/testify/assert"
	boom "github.com/tylertreat/BoomFilters"
	"testing"
)

const testTransaction = `{
    "txid": "512a47f5cfd2e7e22e3d440d3e8c445e41eba55b23ff5f3f696c7f106c22eab3",
    "hash": "512a47f5cfd2e7e22e3d440d3e8c445e41eba55b23ff5f3f696c7f106c22eab3",
    "hex": "0100000001979f9f1033357ab26d71381d40d94069fe2e795aed3b1bc9f28de57c57b278ed030000006a47304402207122f4592d4ddae3214bbafa9b25dd86516bc25698228a3ec1caf8d7b5fabf23022054c1347b421c55cc362227e00c5d18660271b670f2af2484ca1cb31f0ac405f2412102069a1aa13ed2c8f2d1bb7b4c3bee2e8333594ee0c0e2e9188157e8d08b8ceac9ffffffff0123020000000000001976a914a9041707efa4c2edea3e3b93c83330b55c6497d088ac00000000",
    "size": 191,
    "version": 1,
    "locktime": 0,
    "vin": [
        {
            "n": 0,
            "txid": "ed78b2577ce58df2c91b3bed5a792efe6940d9401d38716db27a3533109f9f97",
            "vout": 3,
            "scriptSig": {
                "asm": "304402207122f4592d4ddae3214bbafa9b25dd86516bc25698228a3ec1caf8d7b5fabf23022054c1347b421c55cc362227e00c5d18660271b670f2af2484ca1cb31f0ac405f241 02069a1aa13ed2c8f2d1bb7b4c3bee2e8333594ee0c0e2e9188157e8d08b8ceac9",
                "hex": "47304402207122f4592d4ddae3214bbafa9b25dd86516bc25698228a3ec1caf8d7b5fabf23022054c1347b421c55cc362227e00c5d18660271b670f2af2484ca1cb31f0ac405f2412102069a1aa13ed2c8f2d1bb7b4c3bee2e8333594ee0c0e2e9188157e8d08b8ceac9"
            },
            "sequence": 4294967295,
            "voutDetails": {
                "value": 0.000007,
                "n": 3,
                "scriptPubKey": {
                    "asm": "OP_DUP OP_HASH160 c98e4e2e5ee8ebcd3aad180dd1a95c464e56461f OP_EQUALVERIFY OP_CHECKSIG",
                    "hex": "76a914c98e4e2e5ee8ebcd3aad180dd1a95c464e56461f88ac",
                    "reqSigs": 1,
                    "type": "pubkeyhash",
                    "addresses": [
                        "1KNjJ7PPYT4hjtT19H3dezLgFp6wWSqab5"
                    ],
                    "isTruncated": false
                },
                "scripthash": "8fa0ad201f31ef61df517873040da99b696bceb49c0f519a656c3c51c0e8bcb9"
            }
        }
    ],
    "vout": [
        {
            "value": 0.00000547,
            "n": 0,
            "scriptPubKey": {
                "asm": "OP_DUP OP_HASH160 a9041707efa4c2edea3e3b93c83330b55c6497d0 OP_EQUALVERIFY OP_CHECKSIG",
                "hex": "76a914a9041707efa4c2edea3e3b93c83330b55c6497d088ac",
                "reqSigs": 1,
                "type": "pubkeyhash",
                "addresses": [
                    "1GQg6bWrJyBtdsi34eByLMvu4iRNvJsyxX"
                ],
                "isTruncated": false
            },
            "scripthash": "58fe5b3fbca38ce38e5a1aeba202af70c39c2da7f8899812d5a5e42cf43ece83"
        }
    ],
    "blockhash": "0000000000000000075e74d06bddd5a92cd64c738fe2e4ee71a09fc52bdce984",
    "confirmations": 74,
    "time": 1649037156,
    "blocktime": 1649037156,
    "blockheight": 733617,
    "vincount": 1,
    "voutcount": 1,
    "vinvalue": 0.000007,
    "voutvalue": 0.00000547
}`

func TestBloomProcessor(t *testing.T) {
	type fields struct {
		filter     *boom.StableBloomFilter
		filterType string
	}
	type args struct {
		item   string
		txJson string
		passes bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "valid address locking script with proper tx",
			fields: fields{
				filter:     boom.NewDefaultStableBloomFilter(uint(1000), float64(0.001)),
				filterType: utils.ScriptTypePubKeyHash,
			},
			args: args{
				item:   "76a914a9041707efa4c2edea3e3b93c83330b55c6497d088ac",
				txJson: testTransaction,
				passes: true,
			},
		},
		{
			name: "bad address locking script with tx",
			fields: fields{
				filter:     boom.NewDefaultStableBloomFilter(uint(1000), float64(0.001)),
				filterType: utils.ScriptTypePubKeyHash,
			},
			args: args{
				item:   "efefefefefef",
				txJson: testTransaction,
				passes: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BloomProcessor{
				filter:     tt.fields.filter,
				filterType: tt.fields.filterType,
			}
			m.Add(tt.args.item)
			event := centrifuge.ServerPublishEvent{
				Publication: centrifuge.Publication{
					Data: []byte(tt.args.txJson),
				},
			}
			tx, err := m.FilterMempoolPublishEvent(event)
			assert.NoError(t, err, "%s - mempool filter unexpectedly failed", tt.name)
			if tt.args.passes {
				assert.NotEqualf(t, tx, "", "%s - expected tx to pass processor and didn't", tt.name)
			} else {
				assert.Equalf(t, tx, "", "%s - expected tx to not pass processor and did", tt.name)
			}
		})
	}
}