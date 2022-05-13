package bux

import (
	"context"
	"fmt"
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bk/bip32"
)

// BenchmarkAction_Transaction_recordTransaction will benchmark the method RecordTransaction()
func BenchmarkAction_Transaction_recordTransaction(b *testing.B) {

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, client, xPub, config, err := initBenchmarkData(b)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}

		var draftTransaction *DraftTransaction
		if draftTransaction, err = client.NewTransaction(ctx, xPub.rawXpubKey, config, client.DefaultModelOptions()...); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}

		var xPriv *bip32.ExtendedKey
		if xPriv, err = bip32.NewKeyFromString(testXPriv); err != nil {
			return
		}

		var hexString string
		if hexString, err = draftTransaction.SignInputs(xPriv); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}

		b.StartTimer()
		if _, err = client.RecordTransaction(ctx, xPub.rawXpubKey, hexString, draftTransaction.ID, client.DefaultModelOptions()...); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}
	}
}

// BenchmarkTransaction_newTransaction will benchmark the method newTransaction()
func BenchmarkAction_Transaction_newTransaction(b *testing.B) {

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, client, xPub, config, err := initBenchmarkData(b)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}

		b.StartTimer()
		if _, err = client.NewTransaction(ctx, xPub.rawXpubKey, config, client.DefaultModelOptions()...); err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			b.Fail()
		}
	}
}

func initBenchmarkData(b *testing.B) (context.Context, ClientInterface, *Xpub, *TransactionConfig, error) {
	ctx, client, _ := CreateBenchmarkSQLiteClient(b, false, true,
		WithCustomTaskManager(&taskManagerMockBase{}),
		WithFreeCache(),
		WithIUCDisabled(),
	)

	opts := append(client.DefaultModelOptions(), New())
	xPub, err := client.NewXpub(ctx, testXPub, opts...)
	if err != nil {
		b.Fail()
	}
	destination := newDestination(xPub.GetID(), testLockingScript, opts...)
	if err = destination.Save(ctx); err != nil {
		b.Fail()
	}

	utxo := newUtxo(xPub.GetID(), testTxID, testLockingScript, 1, 122500, opts...)
	if err = utxo.Save(ctx); err != nil {
		b.Fail()
	}
	utxo = newUtxo(xPub.GetID(), testTxID, testLockingScript, 2, 122500, opts...)
	if err = utxo.Save(ctx); err != nil {
		b.Fail()
	}
	utxo = newUtxo(xPub.GetID(), testTxID, testLockingScript, 3, 122500, opts...)
	if err = utxo.Save(ctx); err != nil {
		b.Fail()
	}
	utxo = newUtxo(xPub.GetID(), testTxID, testLockingScript, 4, 122500, opts...)
	if err = utxo.Save(ctx); err != nil {
		b.Fail()
	}

	config := &TransactionConfig{
		FeeUnit: &utils.FeeUnit{
			Satoshis: 5,
			Bytes:    100,
		},
		Outputs: []*TransactionOutput{{
			OpReturn: &OpReturn{
				Map: &MapProtocol{
					App:  "getbux.io",
					Type: "blast",
					Keys: map[string]interface{}{
						"bux": "blasting",
					},
				},
			},
		}},
		ChangeDestinationsStrategy: ChangeStrategyRandom,
		ChangeNumberOfDestinations: 2,
	}

	return ctx, client, xPub, config, err
}
