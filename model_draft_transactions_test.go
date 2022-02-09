package bux

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/libsv/go-bt/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDraftLockingScript = "76a9140692ed78f6988968ce612f620894997cc7edf1ad88ac"
)

// TestDraftTransaction_newDraftTransaction will test the method newDraftTransaction()
func TestDraftTransaction_newDraftTransaction(t *testing.T) {
	t.Parallel()

	t.Run("nil config, panic", func(t *testing.T) {
		assert.Panics(t, func() {
			draftTx := newDraftTransaction(
				testXPub, nil, New(),
			)
			require.NotNil(t, draftTx)
		})
	})

	t.Run("valid config", func(t *testing.T) {
		expires := time.Now().UTC().Add(defaultDraftTxExpiresIn)
		draftTx := newDraftTransaction(
			testXPub, &TransactionConfig{}, New(),
		)
		require.NotNil(t, draftTx)
		assert.NotEqual(t, "", draftTx.ID)
		assert.Equal(t, 64, len(draftTx.ID))
		assert.WithinDurationf(t, expires, draftTx.ExpiresAt, 1*time.Second, "within 1 second")
		assert.Equal(t, DraftStatusDraft, draftTx.Status)
		assert.Equal(t, testXPubID, draftTx.XpubID)
	})
}

// TestDraftTransaction_GetModelName will test the method GetModelName()
func TestDraftTransaction_GetModelName(t *testing.T) {
	t.Parallel()

	t.Run("model name", func(t *testing.T) {
		draftTx := newDraftTransaction(testXPub, &TransactionConfig{}, New())
		require.NotNil(t, draftTx)
		assert.Equal(t, ModelDraftTransaction.String(), draftTx.GetModelName())
	})
}

// TestDraftTransaction_getOutputSatoshis tests getting the output satoshis for the destinations
func TestDraftTransaction_getOutputSatoshis(t *testing.T) {
	t.Run("1 change destination", func(t *testing.T) {
		draftTx := newDraftTransaction(
			testXPub, &TransactionConfig{
				ChangeDestinations: []*Destination{{
					LockingScript: testLockingScript,
				}},
			},
		)
		changSatoshis, err := draftTx.getChangeSatoshis(1000000)
		require.NoError(t, err)
		assert.Len(t, changSatoshis, 1)
		assert.Equal(t, uint64(1000000), changSatoshis[testLockingScript])
	})

	t.Run("2 change destinations", func(t *testing.T) {
		draftTx := newDraftTransaction(
			testXPub, &TransactionConfig{
				ChangeDestinations: []*Destination{{
					LockingScript: testLockingScript,
				}, {
					LockingScript: testTxInScriptPubKey,
				}},
			},
		)
		changSatoshis, err := draftTx.getChangeSatoshis(1000001)
		require.NoError(t, err)
		assert.Len(t, changSatoshis, 2)
		assert.Equal(t, uint64(500000), changSatoshis[testLockingScript])
		assert.Equal(t, uint64(500001), changSatoshis[testTxInScriptPubKey])
	})

	t.Run("3 change destinations - random", func(t *testing.T) {
		draftTx := newDraftTransaction(
			testXPub, &TransactionConfig{
				ChangeDestinationsStrategy: ChangeStrategyRandom,
				ChangeDestinations: []*Destination{{
					LockingScript: testLockingScript,
				}, {
					LockingScript: testTxInScriptPubKey,
				}, {
					LockingScript: testTxScriptPubKey1,
				}},
			},
		)
		satoshis := uint64(1000001)
		changSatoshis, err := draftTx.getChangeSatoshis(satoshis)
		require.NoError(t, err)
		assert.Len(t, changSatoshis, 3)
		totalSatoshis := uint64(0)
		for _, s := range changSatoshis {
			totalSatoshis += s
		}
		assert.Equal(t, totalSatoshis, satoshis)
	})
}

// TestDraftTransaction_setChangeDestinations sets the given of change destinations on the draft transaction
func TestDraftTransaction_setChangeDestinations(t *testing.T) {
	t.Run("1 change destination", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		xPub := newXpub(testXPub, append(client.DefaultModelOptions(), New())...)
		err := xPub.Save(ctx)
		require.NoError(t, err)

		draftTx := newDraftTransaction(testXPub, &TransactionConfig{
			Outputs: []*TransactionOutput{{
				To:       testExternalAddress,
				Satoshis: 1000,
			}},
		}, append(client.DefaultModelOptions(), New())...)

		err = draftTx.setChangeDestinations(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, draftTx.Configuration.ChangeDestinations, 1)
	})

	t.Run("5 change destinations", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		xPub := newXpub(testXPub, append(client.DefaultModelOptions(), New())...)
		err := xPub.Save(ctx)
		require.NoError(t, err)

		draftTx := newDraftTransaction(testXPub, &TransactionConfig{
			Outputs: []*TransactionOutput{{
				To:       testExternalAddress,
				Satoshis: 1000,
			}},
		}, append(client.DefaultModelOptions(), New())...)

		err = draftTx.setChangeDestinations(ctx, 5)
		require.NoError(t, err)
		assert.Len(t, draftTx.Configuration.ChangeDestinations, 5)
	})
}

// TestDraftTransaction_getDraftTransactionID tests getting the draft transaction by draft id
func TestDraftTransaction_getDraftTransactionID(t *testing.T) {

	t.Run("not found", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		draftTx, err := getDraftTransactionID(ctx, testXPubID, testDraftID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Nil(t, draftTx)
	})

	t.Run("found by draft id", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		draftTransaction := newDraftTransaction(testXPub, &TransactionConfig{}, client.DefaultModelOptions()...)
		err := draftTransaction.Save(ctx)
		require.NoError(t, err)

		var draftTx *DraftTransaction
		draftTx, err = getDraftTransactionID(ctx, testXPubID, draftTransaction.ID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Equal(t, 64, len(draftTx.GetID()))
		assert.Equal(t, testXPubID, draftTx.XpubID)
	})
}

// TestDraftTransaction_processOutputs process the outputs of the transaction config
func TestDraftTransaction_processOutputs(t *testing.T) {
	// todo implement test for this using mock paymail client
}

// TestDraftTransaction_createTransaction create a transaction hex
func TestDraftTransaction_createTransaction(t *testing.T) {

	t.Run("empty transaction", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		draftTransaction := newDraftTransaction(testXPub, &TransactionConfig{}, append(client.DefaultModelOptions(), New())...)

		err := draftTransaction.createTransactionHex(ctx)
		require.ErrorIs(t, err, ErrMissingTransactionOutputs)
	})

	t.Run("transaction with no utxos", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		draftTransaction := newDraftTransaction(testXPub, &TransactionConfig{
			Outputs: []*TransactionOutput{{
				To:       testExternalAddress,
				Satoshis: 1000,
			}},
		}, append(client.DefaultModelOptions(), New())...)

		err := draftTransaction.createTransactionHex(ctx)
		require.ErrorIs(t, err, ErrNotEnoughUtxos)
	})

	t.Run("transaction with utxos", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		xPub := newXpub(testXPub, append(client.DefaultModelOptions(), New())...)
		err := xPub.Save(ctx)
		require.NoError(t, err)

		destination := newDestination(testXPubID, testLockingScript,
			append(client.DefaultModelOptions(), New())...)
		err = destination.Save(ctx)
		require.NoError(t, err)

		utxo := newUtxo(testXPubID, testTxID, testLockingScript, 0, 100000,
			append(client.DefaultModelOptions(), New())...)
		err = utxo.Save(ctx)
		require.NoError(t, err)

		draftTransaction := newDraftTransaction(testXPub, &TransactionConfig{
			Outputs: []*TransactionOutput{{
				To:       testExternalAddress,
				Satoshis: 1000,
			}},
		}, append(client.DefaultModelOptions(), New())...)

		err = draftTransaction.createTransactionHex(ctx)
		require.NoError(t, err)
		assert.Equal(t, testXPubID, draftTransaction.XpubID)
		assert.Equal(t, DraftStatusDraft, draftTransaction.Status)

		assert.Equal(t, testXPubID, draftTransaction.Configuration.ChangeDestinations[0].XpubID)
		assert.Equal(t, draftTransaction.ID, draftTransaction.Configuration.ChangeDestinations[0].DraftID)
		assert.Equal(t, uint64(98903), draftTransaction.Configuration.ChangeSatoshis)

		assert.Equal(t, uint64(97), draftTransaction.Configuration.Fee)
		assert.Equal(t, defaultFee, draftTransaction.Configuration.FeeUnit)

		assert.Equal(t, 1, len(draftTransaction.Configuration.Inputs))
		assert.Equal(t, testLockingScript, draftTransaction.Configuration.Inputs[0].ScriptPubKey)
		assert.Equal(t, uint64(100000), draftTransaction.Configuration.Inputs[0].Satoshis)

		assert.Equal(t, 2, len(draftTransaction.Configuration.Outputs))
		assert.Equal(t, uint64(1000), draftTransaction.Configuration.Outputs[0].Satoshis)
		assert.Equal(t, uint64(98903), draftTransaction.Configuration.Outputs[1].Satoshis)
		assert.Equal(t, draftTransaction.Configuration.ChangeDestinations[0].LockingScript, draftTransaction.Configuration.Outputs[1].Scripts[0].Script)

		var btTx *bt.Tx
		btTx, err = bt.NewTxFromString(draftTransaction.Hex)
		require.NoError(t, err)

		assert.Equal(t, 1, len(btTx.Inputs))
		assert.Equal(t, testTxID, hex.EncodeToString(btTx.Inputs[0].PreviousTxID()))
		assert.Equal(t, uint32(0), btTx.Inputs[0].PreviousTxOutIndex)

		assert.Equal(t, 2, len(btTx.Outputs))
		assert.Equal(t, uint64(1000), btTx.Outputs[0].Satoshis)
		assert.Equal(t, draftTransaction.Configuration.Outputs[0].Scripts[0].Script, btTx.Outputs[0].LockingScript.String())

		assert.Equal(t, uint64(98903), btTx.Outputs[1].Satoshis)
		assert.Equal(t, draftTransaction.Configuration.Outputs[1].Scripts[0].Script, btTx.Outputs[1].LockingScript.String())

		var gUtxo *Utxo
		gUtxo, err = getUtxo(ctx, testTxID, 0, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.True(t, gUtxo.DraftID.Valid)
		assert.Equal(t, draftTransaction.ID, gUtxo.DraftID.String)
		assert.True(t, gUtxo.ReservedAt.Valid)
	})

	t.Run("send to all", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		xPub := newXpub(testXPub, append(client.DefaultModelOptions(), New())...)
		err := xPub.Save(ctx)
		require.NoError(t, err)

		destination := newDestination(testXPubID, testLockingScript,
			append(client.DefaultModelOptions(), New())...)
		err = destination.Save(ctx)
		require.NoError(t, err)

		utxo := newUtxo(testXPubID, testTxID, testLockingScript, 0, 100000,
			append(client.DefaultModelOptions(), New())...)
		err = utxo.Save(ctx)
		require.NoError(t, err)

		draftTransaction := newDraftTransaction(testXPub, &TransactionConfig{
			SendAllTo: testExternalAddress,
		}, append(client.DefaultModelOptions(), New())...)

		err = draftTransaction.createTransactionHex(ctx)
		require.NoError(t, err)
		assert.Equal(t, testXPubID, draftTransaction.XpubID)
		assert.Equal(t, DraftStatusDraft, draftTransaction.Status)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.SendAllTo)
		assert.Equal(t, uint64(97), draftTransaction.Configuration.Fee)
		assert.Len(t, draftTransaction.Configuration.Inputs, 1)
		assert.Len(t, draftTransaction.Configuration.Outputs, 1)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.Outputs[0].To)
		assert.Equal(t, uint64(99903), draftTransaction.Configuration.Outputs[0].Satoshis)
		assert.Len(t, draftTransaction.Configuration.Outputs[0].Scripts, 1)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.Outputs[0].Scripts[0].Address)
		assert.Equal(t, uint64(99903), draftTransaction.Configuration.Outputs[0].Scripts[0].Satoshis)
	})

	t.Run("send to all - multiple utxos", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		xPub := newXpub(testXPub, append(client.DefaultModelOptions(), New())...)
		err := xPub.Save(ctx)
		require.NoError(t, err)

		destination := newDestination(testXPubID, testLockingScript,
			append(client.DefaultModelOptions(), New())...)
		err = destination.Save(ctx)
		require.NoError(t, err)

		utxo := newUtxo(testXPubID, testTxID, testLockingScript, 0, 100000,
			append(client.DefaultModelOptions(), New())...)
		err = utxo.Save(ctx)
		require.NoError(t, err)
		utxo = newUtxo(testXPubID, testTxID, testLockingScript, 1, 110000,
			append(client.DefaultModelOptions(), New())...)
		err = utxo.Save(ctx)
		require.NoError(t, err)
		utxo = newUtxo(testXPubID, testTxID, testLockingScript, 2, 130000,
			append(client.DefaultModelOptions(), New())...)
		err = utxo.Save(ctx)
		require.NoError(t, err)

		draftTransaction := newDraftTransaction(testXPub, &TransactionConfig{
			SendAllTo: testExternalAddress,
		}, append(client.DefaultModelOptions(), New())...)

		err = draftTransaction.createTransactionHex(ctx)
		require.NoError(t, err)
		assert.Equal(t, testXPubID, draftTransaction.XpubID)
		assert.Equal(t, DraftStatusDraft, draftTransaction.Status)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.SendAllTo)
		assert.Equal(t, uint64(245), draftTransaction.Configuration.Fee)
		assert.Len(t, draftTransaction.Configuration.Inputs, 3)
		assert.Len(t, draftTransaction.Configuration.Outputs, 1)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.Outputs[0].To)
		assert.Equal(t, uint64(339755), draftTransaction.Configuration.Outputs[0].Satoshis)
		assert.Len(t, draftTransaction.Configuration.Outputs[0].Scripts, 1)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.Outputs[0].Scripts[0].Address)
		assert.Equal(t, uint64(339755), draftTransaction.Configuration.Outputs[0].Scripts[0].Satoshis)
	})

	t.Run("send to all - selected utxos", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		xPub := newXpub(testXPub, append(client.DefaultModelOptions(), New())...)
		err := xPub.Save(ctx)
		require.NoError(t, err)

		destination := newDestination(testXPubID, testLockingScript,
			append(client.DefaultModelOptions(), New())...)
		err = destination.Save(ctx)
		require.NoError(t, err)

		utxo := newUtxo(testXPubID, testTxID, testLockingScript, 0, 100000,
			append(client.DefaultModelOptions(), New())...)
		err = utxo.Save(ctx)
		require.NoError(t, err)
		utxo = newUtxo(testXPubID, testTxID, testLockingScript, 1, 110000,
			append(client.DefaultModelOptions(), New())...)
		err = utxo.Save(ctx)
		require.NoError(t, err)
		utxo = newUtxo(testXPubID, testTxID, testLockingScript, 2, 130000,
			append(client.DefaultModelOptions(), New())...)
		err = utxo.Save(ctx)
		require.NoError(t, err)

		draftTransaction := newDraftTransaction(testXPub, &TransactionConfig{
			SendAllTo: testExternalAddress,
			FromUtxos: []*UtxoPointer{{
				TransactionID: testTxID,
				OutputIndex:   1,
			}, {
				TransactionID: testTxID,
				OutputIndex:   2,
			}},
		}, append(client.DefaultModelOptions(), New())...)

		err = draftTransaction.createTransactionHex(ctx)
		require.NoError(t, err)
		assert.Equal(t, testXPubID, draftTransaction.XpubID)
		assert.Equal(t, DraftStatusDraft, draftTransaction.Status)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.SendAllTo)
		assert.Equal(t, uint64(171), draftTransaction.Configuration.Fee)
		assert.Len(t, draftTransaction.Configuration.Inputs, 2)
		assert.Len(t, draftTransaction.Configuration.Outputs, 1)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.Outputs[0].To)
		assert.Equal(t, uint64(239829), draftTransaction.Configuration.Outputs[0].Satoshis)
		assert.Len(t, draftTransaction.Configuration.Outputs[0].Scripts, 1)
		assert.Equal(t, testExternalAddress, draftTransaction.Configuration.Outputs[0].Scripts[0].Address)
		assert.Equal(t, uint64(239829), draftTransaction.Configuration.Outputs[0].Scripts[0].Satoshis)
	})
}

// TestDraftTransaction_setChangeDestination setting the change destination
func TestDraftTransaction_setChangeDestination(t *testing.T) {
	t.Run("missing xpub", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		draftTransaction := &DraftTransaction{
			Model: *NewBaseModel(
				ModelDraftTransaction,
				append(client.DefaultModelOptions(), WithXPub(testXPub))...,
			),
			Configuration: TransactionConfig{
				ChangeDestinations: nil,
			},
		}

		err := draftTransaction.setChangeDestination(ctx, 100)
		require.ErrorIs(t, err, ErrMissingXpub)
	})

	t.Run("set valid destination", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		xPub := newXpub(testXPub, append(client.DefaultModelOptions(), New())...)
		xPub.NextExternalNum = 121
		xPub.NextInternalNum = 12
		err := xPub.Save(ctx)
		require.NoError(t, err)

		draftTransaction := &DraftTransaction{
			Model: *NewBaseModel(
				ModelDraftTransaction,
				append(client.DefaultModelOptions(), WithXPub(testXPub))...,
			),
			Configuration: TransactionConfig{
				ChangeDestinations: nil,
			},
		}

		err = draftTransaction.setChangeDestination(ctx, 100)
		require.NoError(t, err)
		assert.Equal(t, uint64(100), draftTransaction.Configuration.ChangeSatoshis)
		assert.Equal(t, testXPubID, draftTransaction.Configuration.ChangeDestinations[0].XpubID)
		assert.Equal(t, uint32(1), draftTransaction.Configuration.ChangeDestinations[0].Chain)
		assert.Equal(t, uint32(12), draftTransaction.Configuration.ChangeDestinations[0].Num)
		assert.Equal(t, utils.ScriptTypePubKeyHash, draftTransaction.Configuration.ChangeDestinations[0].Type)
		assert.Equal(t, uint64(100), draftTransaction.Configuration.Outputs[0].Satoshis)
	})
}

// TestDraftTransaction_getInputsFromUtxos getting bt.UTXOs from bux Utxos
func TestDraftTransaction_getInputsFromUtxos(t *testing.T) {
	t.Run("invalid lockingScript", func(t *testing.T) {
		draftTransaction := &DraftTransaction{}

		reservedUtxos := []*Utxo{{
			TransactionID: testTxID,
			OutputIndex:   123,
			Satoshis:      124235,
			ScriptPubKey:  "testLockingScript",
		}}
		inputUtxos, satoshisReserved, err := draftTransaction.getInputsFromUtxos(reservedUtxos)
		require.ErrorIs(t, err, ErrInvalidLockingScript)
		assert.Nil(t, inputUtxos)
		assert.Equal(t, uint64(0), satoshisReserved)
	})

	t.Run("invalid transactionId", func(t *testing.T) {
		draftTransaction := &DraftTransaction{}

		reservedUtxos := []*Utxo{{
			TransactionID: "testTxID",
			OutputIndex:   123,
			Satoshis:      124235,
			ScriptPubKey:  testLockingScript,
		}}
		inputUtxos, satoshisReserved, err := draftTransaction.getInputsFromUtxos(reservedUtxos)
		require.ErrorIs(t, err, ErrInvalidTransactionID)
		assert.Nil(t, inputUtxos)
		assert.Equal(t, uint64(0), satoshisReserved)
	})

	t.Run("get valid", func(t *testing.T) {
		draftTransaction := &DraftTransaction{}

		reservedUtxos := []*Utxo{{
			TransactionID: testTxID,
			OutputIndex:   123,
			Satoshis:      124235,
			ScriptPubKey:  testLockingScript,
		}}
		inputUtxos, satoshisReserved, err := draftTransaction.getInputsFromUtxos(reservedUtxos)
		require.NoError(t, err)
		assert.Equal(t, uint64(124235), satoshisReserved)
		assert.Equal(t, 1, len(*inputUtxos))
		assert.Equal(t, testTxID, hex.EncodeToString((*inputUtxos)[0].TxID))
		assert.Equal(t, uint32(123), (*inputUtxos)[0].Vout)
		assert.Equal(t, testLockingScript, (*inputUtxos)[0].LockingScript.String())
		assert.Equal(t, uint64(124235), (*inputUtxos)[0].Satoshis)
	})

	t.Run("get multi", func(t *testing.T) {
		draftTransaction := &DraftTransaction{}

		reservedUtxos := []*Utxo{{
			TransactionID: testTxID,
			OutputIndex:   124,
			Satoshis:      52313,
			ScriptPubKey:  testLockingScript,
		}, {
			TransactionID: testTxID,
			OutputIndex:   123,
			Satoshis:      124235,
			ScriptPubKey:  testLockingScript,
		}}
		inputUtxos, satoshisReserved, err := draftTransaction.getInputsFromUtxos(reservedUtxos)
		require.NoError(t, err)
		assert.Equal(t, uint64(124235+52313), satoshisReserved)
		assert.Equal(t, 2, len(*inputUtxos))

		assert.Equal(t, testTxID, hex.EncodeToString((*inputUtxos)[0].TxID))
		assert.Equal(t, uint32(124), (*inputUtxos)[0].Vout)
		assert.Equal(t, testLockingScript, (*inputUtxos)[0].LockingScript.String())
		assert.Equal(t, uint64(52313), (*inputUtxos)[0].Satoshis)

		assert.Equal(t, testTxID, hex.EncodeToString((*inputUtxos)[1].TxID))
		assert.Equal(t, uint32(123), (*inputUtxos)[1].Vout)
		assert.Equal(t, testLockingScript, (*inputUtxos)[1].LockingScript.String())
		assert.Equal(t, uint64(124235), (*inputUtxos)[1].Satoshis)
	})
}

// TestDraftTransaction_AfterUpdated after updated tests
func TestDraftTransaction_AfterUpdated(t *testing.T) {
	t.Run("cancel draft - update utxo reservation", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()
		reservationDraftID, _ := utils.RandomHex(32)

		utxo := newUtxo(testXPubID, testTxID, testLockingScript, 0, 100000,
			append(client.DefaultModelOptions(), New())...)
		utxo.DraftID.Valid = true
		utxo.DraftID.String = reservationDraftID
		utxo.ReservedAt.Valid = true
		utxo.ReservedAt.Time = time.Now().UTC()
		err := utxo.Save(ctx)
		require.NoError(t, err)

		var gUtxo *Utxo
		gUtxo, err = getUtxo(ctx, testTxID, 0, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.True(t, gUtxo.DraftID.Valid)
		assert.Equal(t, reservationDraftID, gUtxo.DraftID.String)
		assert.True(t, gUtxo.ReservedAt.Valid)

		draftTransaction := &DraftTransaction{
			Model: *NewBaseModel(
				ModelDraftTransaction,
				client.DefaultModelOptions()...,
			),
			TransactionBase: TransactionBase{ID: reservationDraftID},
			Configuration:   TransactionConfig{},
			Status:          DraftStatusCanceled,
		}

		err = draftTransaction.AfterUpdated(ctx)
		require.NoError(t, err)

		var gUtxo2 *Utxo
		gUtxo2, err = getUtxo(ctx, testTxID, 0, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.False(t, gUtxo2.DraftID.Valid)
		assert.False(t, gUtxo2.ReservedAt.Valid)
	})
}

// TestDraftTransaction_addOutputsToTx will test the method addOutputsToTx()
func TestDraftTransaction_addOutputsToTx(t *testing.T) {
	t.Run("no output", func(t *testing.T) {
		draft := &DraftTransaction{
			Configuration: TransactionConfig{
				Outputs: []*TransactionOutput{{
					Satoshis: 0,
				}},
			},
		}
		tx := bt.NewTx()
		err := draft.addOutputsToTx(tx)
		require.NoError(t, err)
	})

	t.Run("no output", func(t *testing.T) {
		draft := &DraftTransaction{
			Configuration: TransactionConfig{
				Outputs: []*TransactionOutput{{
					Scripts: []*ScriptOutput{{
						Satoshis: 0,
						Script:   testDraftLockingScript,
					}},
				}},
			},
		}
		tx := bt.NewTx()
		err := draft.addOutputsToTx(tx)
		require.ErrorIs(t, err, ErrOutputValueTooLow)
		assert.Len(t, tx.Outputs, 0)
	})

	t.Run("normal address", func(t *testing.T) {
		draft := &DraftTransaction{
			Configuration: TransactionConfig{
				Outputs: []*TransactionOutput{{
					Scripts: []*ScriptOutput{{
						Satoshis: 1000,
						Script:   testDraftLockingScript,
					}},
				}},
			},
		}
		tx := bt.NewTx()
		err := draft.addOutputsToTx(tx)
		require.NoError(t, err)
		assert.Len(t, tx.Outputs, 1)
		assert.Equal(t, uint64(1000), tx.Outputs[0].Satoshis)
		assert.Equal(t, testDraftLockingScript, tx.Outputs[0].LockingScript.String())
	})

	t.Run("op return", func(t *testing.T) {
		draft := &DraftTransaction{
			Configuration: TransactionConfig{
				Outputs: []*TransactionOutput{{
					Scripts: []*ScriptOutput{{
						Satoshis:   0,
						Script:     testDraftLockingScript,
						ScriptType: utils.ScriptTypeNullData,
					}},
				}},
			},
		}
		tx := bt.NewTx()
		err := draft.addOutputsToTx(tx)
		require.NoError(t, err)
		assert.Len(t, tx.Outputs, 1)
		assert.Equal(t, uint64(0), tx.Outputs[0].Satoshis)
		assert.Equal(t, testDraftLockingScript, tx.Outputs[0].LockingScript.String())
	})

	t.Run("op return", func(t *testing.T) {
		draft := &DraftTransaction{
			Configuration: TransactionConfig{
				Outputs: []*TransactionOutput{{
					Scripts: []*ScriptOutput{{
						Satoshis:   1000,
						Script:     testDraftLockingScript,
						ScriptType: utils.ScriptTypeNullData,
					}},
				}},
			},
		}
		tx := bt.NewTx()
		err := draft.addOutputsToTx(tx)
		require.ErrorIs(t, err, ErrInvalidOpReturnOutput)
	})
}

func createDraftTransactionFromHex(hex string, inInfo []interface{}) (*DraftTransaction, *bt.Tx, error) {
	tx, err := bt.NewTxFromString(hex)
	if err != nil {
		return nil, nil, err
	}

	feePaid := uint64(0)

	inputs := make([]*TransactionInput, 0)
	for txIndex := range tx.Inputs {
		in := inInfo[txIndex].(map[string]interface{})
		satoshis := uint64(in["satoshis"].(float64))
		lockingScript := in["locking_script"].(string)
		input := TransactionInput{
			Utxo: Utxo{
				TransactionID: tx.TxID(),
				XpubID:        testXPubID,
				OutputIndex:   uint32(txIndex),
				Satoshis:      satoshis,
				ScriptPubKey:  lockingScript,
				Type:          utils.ScriptTypePubKeyHash,
			},
			Destination: Destination{
				XpubID:        testXPubID,
				LockingScript: lockingScript,
				Type:          utils.ScriptTypePubKeyHash,
				Chain:         0,
				Num:           uint32(txIndex),
				Address:       testExternalAddress,
				DraftID:       testDraftID,
			},
		}
		feePaid += input.Satoshis

		inputs = append(inputs, &input)
	}

	outputs := make([]*TransactionOutput, 0)
	for _, txOutput := range tx.Outputs {
		output := TransactionOutput{
			Satoshis: txOutput.Satoshis,
			Scripts: []*ScriptOutput{{
				Address: testExternalAddress,
				Script:  txOutput.LockingScript.String(),
			}},
			To: testExternalAddress,
		}
		feePaid -= output.Satoshis

		outputs = append(outputs, &output)
	}

	draftConfig := &TransactionConfig{
		ChangeDestinations: []*Destination{{}}, // set to not nil, otherwise will be overwritten when processing
		Fee:                0,
		FeeUnit:            defaultFee,
		Inputs:             inputs,
		Outputs:            outputs,
	}

	return newDraftTransaction(testXPub, draftConfig), tx, nil
}

func TestDraftTransaction_estimateFees(t *testing.T) {
	jsonFile, err := os.Open("./model_draft_transactions_test.json")
	require.NoError(t, err)
	defer func() {
		_ = jsonFile.Close()
	}()

	byteValue, bErr := ioutil.ReadAll(jsonFile)
	require.NoError(t, bErr)

	var testData map[string]interface{}
	err = json.Unmarshal(byteValue, &testData)
	require.NoError(t, err)

	feeUnit := utils.FeeUnit{
		Satoshis: 1,
		Bytes:    2,
	}

	for _, inTx := range testData["rawTransactions"].([]interface{}) {
		in := inTx.(map[string]interface{})
		txID := in["txId"].(string)
		draftTransaction, tx, err2 := createDraftTransactionFromHex(in["hex"].(string), in["inputs"].([]interface{}))
		require.NoError(t, err2)
		assert.Equal(t, txID, tx.TxID())
		assert.IsType(t, DraftTransaction{}, *draftTransaction)
		assert.IsType(t, bt.Tx{}, *tx)

		realFee := uint64(0)
		for _, input := range in["inputs"].([]interface{}) {
			i := input.(map[string]interface{})
			realFee += uint64(i["satoshis"].(float64))
		}
		for _, output := range tx.Outputs {
			realFee -= output.Satoshis
		}

		realSize := uint64(float64(len(in["hex"].(string))) / 2)
		sizeEstimate := draftTransaction.estimateSize()
		feeEstimate := draftTransaction.estimateFee(&feeUnit)
		assert.Greater(t, sizeEstimate, realSize)
		assert.Greater(t, feeEstimate, realFee)
		// fmt.Printf("%s\nSIZE: %d = %d ? => FEE: %d = %d ?\n\n", txID, sizeEstimate, realSize, feeEstimate, realFee)
	}
}

// TestDraftTransaction_RegisterTasks will test the method RegisterTasks()
func TestDraftTransaction_RegisterTasks(t *testing.T) {

	draftCleanupTask := "draft_transaction_clean_up"

	t.Run("testing task: "+draftCleanupTask, func(t *testing.T) {

		_, client, deferMe := CreateTestSQLiteClient(t, false, false)
		defer deferMe()

		assert.Equal(t, 60*time.Second, client.GetTaskPeriod(draftCleanupTask))

		err := client.ModifyTaskPeriod(draftCleanupTask, 10*time.Minute)
		require.NoError(t, err)

		assert.Equal(t, 10*time.Minute, client.GetTaskPeriod(draftCleanupTask))
	})
}
