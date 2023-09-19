package chainstate

import (
	"fmt"
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/centrifugal/centrifuge-go"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	boom "github.com/tylertreat/BoomFilters"
)

const testTransactionHex = `0100000001979f9f1033357ab26d71381d40d94069fe2e795aed3b1bc9f28de57c57b278ed030000006a47304402207122f4592d4ddae3214bbafa9b25dd86516bc25698228a3ec1caf8d7b5fabf23022054c1347b421c55cc362227e00c5d18660271b670f2af2484ca1cb31f0ac405f2412102069a1aa13ed2c8f2d1bb7b4c3bee2e8333594ee0c0e2e9188157e8d08b8ceac9ffffffff0123020000000000001976a914a9041707efa4c2edea3e3b93c83330b55c6497d088ac00000000`

// const testTransactionID = `512a47f5cfd2e7e22e3d440d3e8c445e41eba55b23ff5f3f696c7f106c22eab3`
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
		txJSON string
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
				filterType: utils.P2PKHRegexpString,
			},
			args: args{
				item:   "76a914a9041707efa4c2edea3e3b93c83330b55c6497d088ac",
				txJSON: testTransaction,
				passes: true,
			},
		},
		{
			name: "bad address locking script with tx",
			fields: fields{
				filter:     boom.NewDefaultStableBloomFilter(uint(1000), float64(0.001)),
				filterType: utils.P2PKHRegexpString,
			},
			args: args{
				item:   "efefefefefef",
				txJSON: testTransaction,
				passes: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BloomProcessor{
				filters: map[string]*BloomProcessorFilter{
					tt.fields.filterType: {
						Filter: tt.fields.filter,
						regex:  utils.P2PKHSubstringRegexp,
					},
				},
			}
			err := m.Add(utils.P2PKHRegexpString, tt.args.item)
			require.NoError(t, err)
			event := centrifuge.ServerPublishEvent{
				Publication: centrifuge.Publication{
					Data: []byte(tt.args.txJSON),
				},
			}
			var tx string
			tx, err = m.FilterTransactionPublishEvent(event.Data)
			assert.NoError(t, err, "%s - mempool Filter unexpectedly failed", tt.name)
			if tt.args.passes {
				assert.NotEqualf(t, tx, "", "%s - expected tx to pass processor and didn't", tt.name)
			} else {
				assert.Equalf(t, tx, "", "%s - expected tx to not pass processor and did", tt.name)
			}
		})
	}
}

// BENCHMARKS

func setupBenchmarkData() *BloomProcessor {
	m := NewBloomProcessor(100000, 0.001)

	for i := 0; i < 100000; i++ {
		priv, _ := bitcoin.CreatePrivateKey()
		p2pkh, _ := bscript.NewP2PKHFromPubKeyEC(priv.PubKey())
		_ = m.Add(utils.P2PKHRegexpString, p2pkh.String())
	}

	return m
}

func getEvent(i int) centrifuge.ServerPublishEvent {
	return centrifuge.ServerPublishEvent{
		Publication: centrifuge.Publication{
			Data: []byte(testTransactionHex + string(rune(i))),
		},
	}
}

// always record the result of the function call to prevent the compiler eliminating it
var txResult []int

func BenchmarkBloomProcessor(b *testing.B) {
	m := setupBenchmarkData()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tx, _ := m.FilterTransactionPublishEvent(getEvent(i).Data)
		txResult = append(txResult, len(tx))
	}
}

func TestBloomProcessor_FilterMempoolTx(t *testing.T) {
	txHex := "01000000013c78aba57467f8cdb6270f171ab0c67df0ecfe65dfce7a0aeeaa5774f6e0f6e3040000006a473044022064b81316db2fe23c598fd6ace8e11ce5668f1006bbdf1a660acf8773852ece29022014a31aee48f007fb4c78014a01e9595dd81f6cae39591901921f30465c2149764121022d6799224389f3764f7610080f06745d87c7296d27db034a98923242cdb280cfffffffff0750c30000000000001976a91481c80d970b24fb03362be1c65145544892cebe5688ac8d160000000000001976a91447f73f8a7807d8ab2e321c76e321d69598343fb188ac400d0300000000001976a914b1a1ffb7a9a78aae58940c73cd2a6c7c170c44f188ac400d0300000000001976a914b762ca8682f2110a540ea8795a2e72c608b6606288ac204e0000000000001976a914455525dda77e409082c691deee061c1fb6b0082088ac204e0000000000001976a914087e1729fe365716080a452bec0103bd7204021c88acbc0f0000000000001976a914f063c2a5c290c525abba30d8085cd9c661ba091288ac00000000"

	type fields struct {
		maxCells          uint
		falsePositiveRate float64
	}
	type args struct {
		txHex string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Real tx",
			fields: fields{
				maxCells:          100,
				falsePositiveRate: 0.01,
			},
			args: args{
				txHex: txHex,
			},
			want:    txHex,
			wantErr: assert.NoError,
		},
		{
			name: "No match",
			fields: fields{
				maxCells:          100,
				falsePositiveRate: 0.01,
			},
			args: args{
				txHex: "test string",
			},
			want:    "",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewBloomProcessor(
				tt.fields.maxCells,
				tt.fields.falsePositiveRate,
			)
			err := p.Add(utils.P2PKHRegexpString, "76a91481c80d970b24fb03362be1c65145544892cebe5688ac")
			require.NoError(t, err)

			var got string
			got, err = p.FilterTransaction(tt.args.txHex)
			if !tt.wantErr(t, err, fmt.Sprintf("FilterMempoolTx(%v)", tt.args.txHex)) {
				return
			}
			assert.Equalf(t, tt.want, got, "FilterMempoolTx(%v)", tt.args.txHex)
		})
	}
}
