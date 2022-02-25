package bux

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/BuxOrg/bux/datastore"
	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/bip32"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testTxHex            = "020000000165bb8d2733298b2d3b441a871868d6323c5392facf0d3eced3a6c6a17dc84c10000000006a473044022057b101e9a017cdcc333ef66a4a1e78720ae15adf7d1be9c33abec0fe56bc849d022013daa203095522039fadaba99e567ec3cf8615861d3b7258d5399c9f1f4ace8f412103b9c72aebee5636664b519e5f7264c78614f1e57fa4097ae83a3012a967b1c4b9ffffffff03e0930400000000001976a91413473d21dc9e1fb392f05a028b447b165a052d4d88acf9020000000000001976a91455decebedd9a6c2c2d32cf0ee77e2640c3955d3488ac00000000000000000c006a09446f7457616c6c657400000000"
	testTxID             = "1b52eac9d1eb0adf3ce6a56dee1c4768780b8126e288aca65dd1db32f173b853"
	testTxID2            = "104cc87da1c6a6d3ce3e0dcffa92533c32d66818871a443b2d8b2933278dbb65"
	testTxInID           = "9b0495704e23e4b3bef3682c6a5c40abccc32a3e6b7b01ae3295e93a9d3a0482"
	testTxInHex          = "020000000189fbccca3a5e2bfc8a161bf7f54e8cb5898e296ae8c23b620b89ed570711f931000000006a47304402204e94380ae4d27f8bb9b40dd9944b4fea532d5fe12cf62c1994a6a495c81490f202204aab42f8f1b15259a032e58a3810fbbfd691771b92317f8a12a0da84761a400641210382229c0295e4d63ee54c541eba40be2963f0e80489b7da34e022d513a723181fffffffff0259970400000000001976a914e069bd2e2fe3ea702c40d5e65b491b734c01686788ac00000000000000000c006a09446f7457616c6c657400000000"
	testTxInScriptPubKey = "76a914e069bd2e2fe3ea702c40d5e65b491b734c01686788ac"
	testTxScriptPubKey1  = "76a91413473d21dc9e1fb392f05a028b447b165a052d4d88ac"
	testTxScriptPubKey2  = "76a91455decebedd9a6c2c2d32cf0ee77e2640c3955d3488ac"
	testTxScriptSigID    = "104cc87da1c6a6d3ce3e0dcffa92533c32d66818871a443b2d8b2933278dbb65"
	testTxScriptSigOut   = "76a914e069bd2e2fe3ea702c40d5e65b491b734c01686788ac"
	// testTxInID           = "9b0495704e23e4b3bef3682c6a5c40abccc32a3e6b7b01ae3295e93a9d3a0482"
)

type transactionServiceMock struct {
	destinations map[string]*Destination
	utxos        map[string]map[uint32]*Utxo
}

func (x transactionServiceMock) getDestinationByLockingScript(_ context.Context, lockingScript string, _ ...ModelOps) (*Destination, error) {
	return x.destinations[lockingScript], nil
}

func (x transactionServiceMock) getUtxo(_ context.Context, txID string, index uint32, _ ...ModelOps) (*Utxo, error) {
	return x.utxos[txID][index], nil
}

// TestTransaction_newTransaction will test the method newTransaction()
func TestTransaction_newTransaction(t *testing.T) {
	t.Parallel()

	t.Run("New transaction model", func(t *testing.T) {
		transaction := newTransaction(testTxHex, New())
		require.NotNil(t, transaction)
		assert.IsType(t, Transaction{}, *transaction)
		assert.Equal(t, ModelTransaction.String(), transaction.GetModelName())
		assert.Equal(t, testTxID, transaction.ID)
		assert.Equal(t, testTxID, transaction.GetID())
		assert.Equal(t, true, transaction.IsNew())
	})

	t.Run("New transaction model - no hex, no options", func(t *testing.T) {
		transaction := newTransaction("")
		require.NotNil(t, transaction)
		assert.IsType(t, Transaction{}, *transaction)
		assert.Equal(t, ModelTransaction.String(), transaction.GetModelName())
		assert.Equal(t, "", transaction.ID)
		assert.Equal(t, "", transaction.GetID())
		assert.Equal(t, false, transaction.IsNew())
	})
}

// TestTransaction_newTransactionWithDraftID will test the method newTransactionWithDraftID()
func TestTransaction_newTransactionWithDraftID(t *testing.T) {
	t.Parallel()

	t.Run("New transaction model", func(t *testing.T) {
		transaction := newTransactionWithDraftID(testTxHex, testDraftID, New())
		require.NotNil(t, transaction)
		assert.IsType(t, Transaction{}, *transaction)
		assert.Equal(t, ModelTransaction.String(), transaction.GetModelName())
		assert.Equal(t, testTxID, transaction.ID)
		assert.Equal(t, testDraftID, transaction.DraftID)
		assert.Equal(t, testTxID, transaction.GetID())
		assert.Equal(t, true, transaction.IsNew())
	})

	t.Run("New transaction model - no hex, no options", func(t *testing.T) {
		transaction := newTransactionWithDraftID("", "")
		require.NotNil(t, transaction)
		assert.IsType(t, Transaction{}, *transaction)
		assert.Equal(t, ModelTransaction.String(), transaction.GetModelName())
		assert.Equal(t, "", transaction.ID)
		assert.Equal(t, "", transaction.DraftID)
		assert.Equal(t, "", transaction.GetID())
		assert.Equal(t, false, transaction.IsNew())
	})
}

// TestTransaction_getTransactionByID will test the method getTransactionByID()
func TestTransaction_getTransactionByID(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		transaction, err := getTransactionByID(ctx, testXPubID, testTxID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Nil(t, transaction)
	})

	t.Run("found tx", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, true)
		defer deferMe()
		opts := client.DefaultModelOptions()
		tx := newTransaction(testTxHex, append(opts, New())...)
		txErr := tx.Save(ctx)
		require.NoError(t, txErr)

		transaction, err := getTransactionByID(ctx, testXPubID, testTxID, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.NotNil(t, transaction)
		assert.Equal(t, testTxID, transaction.ID)
		assert.Equal(t, testTxHex, transaction.Hex)
		assert.Nil(t, transaction.XpubInIDs)
		assert.Nil(t, transaction.XpubOutIDs)
	})
}

// TestTransaction_getTransactionsByXpubID will test the method getTransactionsByXpubID()
func TestTransaction_getTransactionsByXpubID(t *testing.T) {
	t.Run("tx not found", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}))
		defer deferMe()
		transactions, err := getTransactionsByXpubID(ctx, testXPub, nil, nil, 0, 0, client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Nil(t, transactions)
	})

	t.Run("tx found", func(t *testing.T) {
		ctx, client, _ := CreateTestSQLiteClient(t, true, true)
		opts := client.DefaultModelOptions()
		tx := newTransaction(testTxHex, append(opts, New())...)
		tx.XpubInIDs = append(tx.XpubInIDs, testXPubID)
		txErr := tx.Save(ctx)
		require.NoError(t, txErr)

		transactions, err := getTransactionsByXpubID(ctx, testXPub, nil, nil, 0, 0, opts...)
		require.NoError(t, err)
		require.NotNil(t, transactions)
		require.Len(t, transactions, 1)
		assert.Equal(t, testTxID, transactions[0].ID)
		assert.Equal(t, testTxHex, transactions[0].Hex)
		assert.Equal(t, testXPubID, transactions[0].XpubInIDs[0])
		assert.Nil(t, transactions[0].XpubOutIDs)
	})
}

// TestTransaction_BeforeCreating will test the method BeforeCreating()
func TestTransaction_BeforeCreating(t *testing.T) {
	// t.Parallel()

	t.Run("incorrect transaction hex", func(t *testing.T) {
		transaction := newTransaction("test")
		err := transaction.BeforeCreating(context.Background())
		assert.Error(t, err)
	})

	t.Run("no transaction hex", func(t *testing.T) {
		transaction := newTransaction("")
		err := transaction.BeforeCreating(context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, ErrMissingFieldHex, err)
	})
}

// TestTransaction_BeforeCreating will test the method BeforeCreating()
func (ts *EmbeddedDBTestSuite) TestTransaction_BeforeCreating() {
	ts.T().Run("[sqlite] [in-memory] - valid transaction", func(t *testing.T) {
		tc := ts.genericDBClient(t, datastore.SQLite, true)
		defer tc.Close(tc.ctx)

		transaction := newTransaction(testTxHex, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, transaction)

		err := transaction.BeforeCreating(tc.ctx)
		require.NoError(t, err)
	})
}

// TestTransaction_GetID will test the method GetID()
func TestTransaction_GetID(t *testing.T) {
	t.Parallel()

	t.Run("no id", func(t *testing.T) {
		transaction := newTransaction("")
		require.NotNil(t, transaction)
		assert.Equal(t, "", transaction.GetID())
	})

	t.Run("valid id", func(t *testing.T) {
		transaction := newTransaction(testTxHex)
		require.NotNil(t, transaction)
		assert.Equal(t, testTxID, transaction.GetID())
	})
}

// TestTransaction_GetModelName will test the method GetModelName()
func TestTransaction_GetModelName(t *testing.T) {
	t.Parallel()

	t.Run("model name", func(t *testing.T) {
		transaction := newTransaction("")
		assert.Equal(t, ModelTransaction.String(), transaction.GetModelName())
	})
}

// TestTransaction_processOutputs will test the method processConfigOutputs()
func TestTransaction_processOutputs(t *testing.T) {
	// t.Parallel()

	t.Run("no outputs", func(t *testing.T) {
		transaction := newTransaction(testTxHex, New())
		require.NotNil(t, transaction)

		transaction.transactionService = transactionServiceMock{}

		ctx := context.Background()
		err := transaction.processOutputs(ctx)
		require.NoError(t, err)
		assert.Nil(t, transaction.utxos)
		assert.Nil(t, transaction.XpubOutIDs)
	})

	t.Run("no outputs", func(t *testing.T) {
		transaction := newTransaction(testTxHex, New())
		require.NotNil(t, transaction)

		transaction.transactionService = transactionServiceMock{
			destinations: map[string]*Destination{
				"76a91413473d21dc9e1fb392f05a028b447b165a052d4d88ac": {
					Model:  Model{name: ModelDestination},
					XpubID: "test-xpub-id",
				},
			},
		}

		ctx := context.Background()
		err := transaction.processOutputs(ctx)
		require.NoError(t, err)
		require.NotNil(t, transaction.utxos)
		assert.IsType(t, Utxo{}, transaction.utxos[0])
		assert.Equal(t, testTxID, transaction.utxos[0].TransactionID)
		assert.Equal(t, "test-xpub-id", transaction.utxos[0].XpubID)
		assert.Equal(t, "test-xpub-id", transaction.XpubOutIDs[0])

		childModels := transaction.ChildModels()
		assert.Len(t, childModels, 1)
		assert.Equal(t, "utxo", childModels[0].Name())
	})
}

// TestTransaction_processInputs will test the method processInputs()
func TestTransaction_processInputs(t *testing.T) {
	// t.Parallel()

	t.Run("no utxo", func(t *testing.T) {
		transaction := newTransaction(testTxHex, New())
		require.NotNil(t, transaction)

		transaction.transactionService = transactionServiceMock{}

		ctx := context.Background()
		err := transaction.processInputs(ctx)
		require.NoError(t, err)
		assert.Nil(t, transaction.utxos)
		assert.Nil(t, transaction.XpubInIDs)
	})

	t.Run("got utxo", func(t *testing.T) {
		transaction := newTransaction(testTxHex, New())
		require.NotNil(t, transaction)

		transaction.draftTransaction = &DraftTransaction{
			TransactionBase: TransactionBase{ID: testDraftID},
		}
		transaction.transactionService = transactionServiceMock{
			utxos: map[string]map[uint32]*Utxo{
				testTxID2: {
					uint32(0): {
						Model:         Model{name: ModelUtxo},
						TransactionID: testTxID2,
						OutputIndex:   0,
						XpubID:        "test-xpub-id",
						DraftID: utils.NullString{NullString: sql.NullString{
							Valid:  true,
							String: testDraftID,
						}},
					},
				},
			},
		}

		ctx := context.Background()
		err := transaction.processInputs(ctx)
		require.NoError(t, err)
		require.NotNil(t, transaction.utxos)
		assert.IsType(t, Utxo{}, transaction.utxos[0])
		assert.Equal(t, testTxID2, transaction.utxos[0].TransactionID)
		assert.True(t, transaction.utxos[0].SpendingTxID.Valid)
		assert.Equal(t, testTxID, transaction.utxos[0].SpendingTxID.String)
		assert.Equal(t, "test-xpub-id", transaction.utxos[0].XpubID)
		assert.Equal(t, "test-xpub-id", transaction.XpubInIDs[0])

		childModels := transaction.ChildModels()
		assert.Len(t, childModels, 1)
		assert.Equal(t, "utxo", childModels[0].Name())
	})

	t.Run("spent utxo", func(t *testing.T) {
		transaction := newTransaction(testTxHex, New())
		require.NotNil(t, transaction)

		transaction.draftTransaction = &DraftTransaction{}
		transaction.transactionService = transactionServiceMock{
			utxos: map[string]map[uint32]*Utxo{
				testTxID2: {
					uint32(0): {
						Model:         Model{name: ModelUtxo},
						TransactionID: testTxID2,
						OutputIndex:   0,
						XpubID:        "test-xpub-id",
						SpendingTxID: utils.NullString{NullString: sql.NullString{
							Valid:  true,
							String: testTxID,
						}},
						DraftID: utils.NullString{NullString: sql.NullString{
							Valid:  true,
							String: testDraftID2,
						}},
					},
				},
			},
		}

		ctx := context.Background()
		err := transaction.processInputs(ctx)
		require.ErrorIs(t, err, ErrUtxoAlreadySpent)
	})

	t.Run("not reserved utxo", func(t *testing.T) {
		transaction := newTransaction(testTxHex, New())
		require.NotNil(t, transaction)

		transaction.draftTransaction = &DraftTransaction{
			TransactionBase: TransactionBase{ID: testDraftID},
		}
		transaction.transactionService = transactionServiceMock{
			utxos: map[string]map[uint32]*Utxo{
				testTxID2: {
					uint32(0): {
						Model:         Model{name: ModelUtxo},
						TransactionID: testTxID2,
						OutputIndex:   0,
						XpubID:        "test-xpub-id",
					},
				},
			},
		}

		ctx := context.Background()
		err := transaction.processInputs(ctx)
		require.ErrorIs(t, err, ErrUtxoNotReserved)
	})

	t.Run("incorrect reservation ID of utxo", func(t *testing.T) {
		transaction := newTransaction(testTxHex, New())
		require.NotNil(t, transaction)

		transaction.draftTransaction = &DraftTransaction{
			TransactionBase: TransactionBase{ID: testDraftID},
		}
		transaction.transactionService = transactionServiceMock{
			utxos: map[string]map[uint32]*Utxo{
				testTxID2: {
					uint32(0): {
						Model:         Model{name: ModelUtxo},
						TransactionID: testTxID2,
						OutputIndex:   0,
						XpubID:        "test-xpub-id",
						DraftID: utils.NullString{NullString: sql.NullString{
							Valid:  true,
							String: testDraftID2,
						}},
					},
				},
			},
		}

		ctx := context.Background()
		err := transaction.processInputs(ctx)
		require.ErrorIs(t, err, ErrDraftIDMismatch)
	})

	t.Run("inputUtxoChecksOff", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(t, false, false, WithCustomTaskManager(&taskManagerMockBase{}), WithIUCDisabled())
		defer deferMe()

		transaction := newTransaction(testTxHex, append(client.DefaultModelOptions(), New())...)
		require.NotNil(t, transaction)

		transaction.draftTransaction = &DraftTransaction{
			TransactionBase: TransactionBase{ID: testDraftID},
		}
		transaction.transactionService = transactionServiceMock{
			utxos: map[string]map[uint32]*Utxo{
				testTxID2: {
					uint32(0): {
						Model:         Model{name: ModelUtxo},
						TransactionID: testTxID2,
						OutputIndex:   0,
						XpubID:        "test-xpub-id",
					},
				},
			},
		}

		err := transaction.processInputs(ctx)
		require.NoError(t, err)
	})
}

func TestTransaction_Display(t *testing.T) {
	t.Run("display without xpub data", func(t *testing.T) {
		tx := Transaction{
			Model:  Model{},
			xPubID: testXPubID,
		}

		displayTx := tx.Display().(*Transaction)
		assert.Nil(t, displayTx.Metadata)
		assert.Equal(t, int64(0), displayTx.OutputValue)
		assert.Nil(t, displayTx.XpubInIDs)
		assert.Nil(t, displayTx.XpubOutIDs)
		assert.Nil(t, displayTx.XpubMetadata)
		assert.Nil(t, displayTx.XpubOutputValue)
	})

	t.Run("display with xpub data", func(t *testing.T) {
		tx := Transaction{
			TransactionBase: TransactionBase{
				ID:  testTxID,
				Hex: "hex",
			},
			Model:           Model{},
			XpubInIDs:       IDs{testXPubID},
			XpubOutIDs:      nil,
			BlockHash:       "hash",
			BlockHeight:     123,
			Fee:             321,
			NumberOfInputs:  1,
			NumberOfOutputs: 2,
			DraftID:         testDraftID,
			TotalValue:      123499,
			XpubMetadata: XpubMetadata{
				testXPubID: Metadata{
					"test-key": "test-value",
				},
			},
			OutputValue: 12,
			XpubOutputValue: XpubOutputValue{
				testXPubID: 123499,
			},
			xPubID: testXPubID,
		}

		displayTx := tx.Display().(*Transaction)
		assert.Equal(t, Metadata{"test-key": "test-value"}, displayTx.Metadata)
		assert.Equal(t, int64(123499), displayTx.OutputValue)
		assert.Equal(t, TransactionDirectionIn, displayTx.Direction)
		assert.Nil(t, displayTx.XpubInIDs)
		assert.Nil(t, displayTx.XpubOutIDs)
		assert.Nil(t, displayTx.XpubMetadata)
		assert.Nil(t, displayTx.XpubOutputValue)
	})
}

// TestTransaction_Save will test the method Save()
func (ts *EmbeddedDBTestSuite) TestTransaction_Save() {
	parsedTx, errP := bt.NewTxFromString(testTxHex)
	require.NoError(ts.T(), errP)
	require.NotNil(ts.T(), parsedTx)

	var parsedInTx *bt.Tx
	parsedInTx, errP = bt.NewTxFromString(testTxInHex)
	require.NoError(ts.T(), errP)
	require.NotNil(ts.T(), parsedInTx)

	ts.T().Run("[sqlite] [in-memory] - Save transaction", func(t *testing.T) {
		tc := ts.genericDBClient(t, datastore.SQLite, false)
		defer tc.Close(tc.ctx)

		transaction := newTransaction(testTxHex, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, transaction)

		err := transaction.Save(tc.ctx)
		require.NoError(t, err)

		var transaction2 *Transaction
		transaction2, err = tc.client.GetTransaction(tc.ctx, testXPubID, testTxID)
		require.NoError(t, err)
		require.NotNil(t, transaction2)
		assert.Equal(t, transaction2.ID, testTxID)

		// no utxos should have been saved, we don't recognize any of the destinations
		var utxo *Utxo
		utxo, err = getUtxo(tc.ctx, transaction.ID, 0, tc.client.DefaultModelOptions()...)
		require.NoError(t, err)
		assert.Nil(t, utxo)
	})

	ts.T().Run("[sqlite] [in-memory] - Save transaction - with utxos & outputs", func(t *testing.T) {
		tc := ts.genericDBClient(t, datastore.SQLite, false)
		defer tc.Close(tc.ctx)

		_, xPub, _ := CreateNewXPub(tc.ctx, t, tc.client)
		require.NotNil(t, xPub)

		_, xPub2, _ := CreateNewXPub(tc.ctx, t, tc.client)
		require.NotNil(t, xPub2)

		// NOTE: these are fake destinations, might want to replace with actual real data / methods

		// fake existing destinations, to generate utxos
		ls := parsedTx.Outputs[0].LockingScript
		destination := newDestination(xPub.GetID(), ls.String(), append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, destination)

		err := destination.Save(tc.ctx)
		require.NoError(t, err)

		ls2 := parsedTx.Outputs[1].LockingScript
		destination2 := newDestination(xPub2.GetID(), ls2.String(), append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, destination2)

		err = destination2.Save(tc.ctx)
		require.NoError(t, err)

		transaction := newTransaction(testTxHex, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, transaction)

		err = transaction.Save(tc.ctx)
		require.NoError(t, err)

		// check whether the XpubOutIDs were set properly
		var transaction2 *Transaction
		transaction2, err = tc.client.GetTransaction(tc.ctx, testXPubID, testTxID)
		require.NoError(t, err)
		require.NotNil(t, transaction2)
		assert.Equal(t, xPub.GetID(), transaction2.XpubOutIDs[0])
		assert.Equal(t, xPub2.GetID(), transaction2.XpubOutIDs[1])

		// utxos should have been saved for our fake destinations
		var utxo *Utxo
		utxo, err = getUtxo(tc.ctx, transaction.ID, 0, tc.client.DefaultModelOptions()...)
		require.NoError(t, err)
		require.NotNil(t, utxo)
		assert.Equal(t, xPub.GetID(), utxo.XpubID)
		assert.Equal(t, bscript.ScriptTypePubKeyHash, utxo.Type)
		assert.Equal(t, testTxScriptPubKey1, utxo.ScriptPubKey)
		assert.Empty(t, utxo.DraftID)
		assert.Empty(t, utxo.SpendingTxID)

		var utxo2 *Utxo
		utxo2, err = getUtxo(tc.ctx, transaction.ID, 1, tc.client.DefaultModelOptions()...)
		assert.Nil(t, err)
		assert.Equal(t, xPub2.GetID(), utxo2.XpubID)
		assert.Equal(t, bscript.ScriptTypePubKeyHash, utxo2.Type)
		assert.Equal(t, testTxScriptPubKey2, utxo2.ScriptPubKey)
		assert.Empty(t, utxo2.DraftID)
		assert.Empty(t, utxo2.SpendingTxID)
	})

	ts.T().Run("[sqlite] [in-memory] - Save transaction - with inputs", func(t *testing.T) {
		tc := ts.genericDBClient(t, datastore.SQLite, false)
		defer tc.Close(tc.ctx)

		_, xPub, _ := CreateNewXPub(tc.ctx, t, tc.client)
		require.NotNil(t, xPub)

		// NOTE: these are fake destinations, might want to replace with actual real data / methods

		// create a fake destination for our IN transaction
		ls := parsedInTx.Outputs[0].LockingScript
		destination := newDestination(xPub.GetID(), ls.String(), append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, destination)

		err := destination.Save(tc.ctx)
		require.NoError(t, err)

		// add the IN transaction
		transactionIn := newTransaction(testTxInHex, append(tc.client.DefaultModelOptions(), New())...)
		require.NotNil(t, transactionIn)

		err = transactionIn.Save(tc.ctx)
		require.NoError(t, err)

		var utxoIn *Utxo
		utxoIn, err = getUtxo(tc.ctx, transactionIn.ID, 0, tc.client.DefaultModelOptions()...)
		require.NotNil(t, utxoIn)
		require.NoError(t, err)
		assert.Equal(t, xPub.GetID(), utxoIn.XpubID)
		assert.Equal(t, bscript.ScriptTypePubKeyHash, utxoIn.Type)
		assert.Equal(t, testTxInScriptPubKey, utxoIn.ScriptPubKey)
		assert.Empty(t, utxoIn.SpendingTxID)

		draftConfig := &TransactionConfig{
			Outputs: []*TransactionOutput{{
				Satoshis: 100,
				To:       testExternalAddress,
			}},
		}
		draftTransaction := newDraftTransaction(
			xPub.rawXpubKey, draftConfig, append(tc.client.DefaultModelOptions(), New())...,
		)
		err = draftTransaction.Save(tc.ctx)
		require.NoError(t, err)

		// this transaction should spend the utxo of the IN transaction
		transaction := newTransaction(testTxHex,
			append(tc.client.DefaultModelOptions(), WithXPub(xPub.rawXpubKey), New())...)
		require.NotNil(t, transactionIn)
		transaction.DraftID = draftTransaction.ID

		err = transaction.Save(tc.ctx)
		require.NoError(t, err)

		// check whether the XpubInIDs were set properly
		var transaction2 *Transaction
		transaction2, err = tc.client.GetTransaction(tc.ctx, testXPubID, testTxID)
		require.NotNil(t, transaction2)
		require.NoError(t, err)
		assert.Equal(t, xPub.GetID(), transaction2.XpubInIDs[0])

		// Get the utxo for the IN transaction and make sure it is marked as spent
		var utxo *Utxo
		utxo, err = getUtxo(tc.ctx, transactionIn.ID, 0, tc.client.DefaultModelOptions()...)
		require.NotNil(t, transaction2)
		require.NoError(t, err)
		assert.Equal(t, testTxInID, utxo.ID)
		assert.True(t, utxo.SpendingTxID.Valid)
		assert.Equal(t, utxo.SpendingTxID.String, testTxID)
	})
}

// BenchmarkTransaction_newTransaction will benchmark the method newTransaction()
func BenchmarkTransaction_newTransaction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = newTransaction(testTxHex, New())
	}
}

// TestEndToEndTransaction will test full end-to-end transaction use cases
func TestEndToEndTransaction(t *testing.T) {
	t.Run("one key, funding tx, create standard tx", func(t *testing.T) {
		ctx, client, deferMe := CreateTestSQLiteClient(
			t, false, true,
			WithCustomChainstate(&chainStateEverythingOnChain{}),
		)
		defer deferMe()

		// Get new random key
		masterKey, xPub, rawXPub := CreateNewXPub(ctx, t, client)
		require.NotNil(t, xPub)

		opts := append(
			client.DefaultModelOptions(),
			WithMetadatas(map[string]interface{}{
				testMetadataKey: testMetadataValue,
			}),
		)

		// Create two destinations (to receive the fake funding transaction)
		var err error
		destinations := make([]*Destination, 2)
		destinations[0], err = client.NewDestination(
			ctx, rawXPub, utils.ChainExternal, utils.ScriptTypePubKeyHash,
			opts...,
		)
		require.NoError(t, err)

		opts2 := append(
			client.DefaultModelOptions(),
			WithMetadatas(map[string]interface{}{
				testMetadataKey + "_2": testMetadataValue + "_2",
			}),
		)

		destinations[1], err = client.NewDestination(
			ctx, rawXPub, utils.ChainExternal, utils.ScriptTypePubKeyHash,
			opts2...,
		)
		require.NoError(t, err)

		// Create the fake funding transaction using the given destinations
		tx := CreateFakeFundingTransaction(t, masterKey, destinations, 10000)

		// Register funding transaction in bux
		var transaction *Transaction
		transaction, err = client.RecordTransaction(ctx, rawXPub, tx,
			"",
			WithMetadata("custom_tag", "custom_value"),
		)
		require.NoError(t, err)
		require.NotNil(t, transaction)
		require.Equal(t, "", transaction.DraftID)
		require.Equal(t, SyncStatusProcessing, transaction.Status)

		// Get the transaction (now after processing)
		transaction, err = client.GetTransaction(ctx, rawXPub, transaction.ID)
		require.NoError(t, err)
		require.NotNil(t, transaction)
		require.Equal(t, SyncStatusComplete, transaction.Status)
		assert.Equal(t, uint32(2), transaction.NumberOfOutputs)
		require.Equal(t, uint64(20000), transaction.TotalValue, transaction.TotalValue)

		assert.Equal(t, uint64(0), transaction.Fee) // fee is zero, we do not have the inputs
		assert.Equal(t, uint32(1), transaction.NumberOfInputs)

		opts3 := client.DefaultModelOptions()

		// Create new draft transaction to an external address
		var draftTransaction *DraftTransaction
		draftTransaction, err = client.NewTransaction(ctx, rawXPub, &TransactionConfig{
			Sync:      &SyncConfig{Broadcast: false, SyncOnChain: false},
			ExpiresIn: 5 * time.Second,
			Outputs: []*TransactionOutput{
				{
					Satoshis: 5000,
					To:       "1LVvLTwaHc7WzKsS5naRov7j3bqQctPPND",
				},
			},
		}, opts3...)
		require.NoError(t, err)
		require.NotNil(t, draftTransaction)

		// Set vars
		var (
			chainKey    *bip32.ExtendedKey
			destination *Destination
			ls          *bscript.Script
			numKey      *bip32.ExtendedKey
			privateKey  *bec.PrivateKey
			txDraft     *bt.Tx
			utxo        *Utxo
		)

		//
		// This should be done by the user's wallet - import the tx and sign the inputs
		// The utxos and destinations will be returned to the request for a template transaction
		//
		txDraft, err = bt.NewTxFromString(draftTransaction.Hex)
		require.NoError(t, err)

		// Loop and sign all inputs
		for index, input := range txDraft.Inputs {
			utxo, err = client.GetUtxo(ctx, rawXPub, input.PreviousTxIDStr(), input.PreviousTxOutIndex)
			require.NoError(t, err)
			require.NotNil(t, utxo)

			destination, err = client.GetDestinationByLockingScript(ctx, rawXPub, utxo.ScriptPubKey)
			require.NoError(t, err)
			require.NotNil(t, destination)

			ls, err = bscript.NewFromHexString(destination.LockingScript)
			require.NoError(t, err)
			require.NotNil(t, ls)

			txDraft.Inputs[index].PreviousTxScript = ls

			chainKey, err = masterKey.Child(destination.Chain)
			require.NoError(t, err)
			require.NotNil(t, chainKey)

			numKey, err = chainKey.Child(destination.Num)
			require.NoError(t, err)
			require.NotNil(t, numKey)

			privateKey, err = bitcoin.GetPrivateKeyFromHDKey(numKey)
			require.NoError(t, err)
			require.NotNil(t, privateKey)

			err = txDraft.InsertInputUnlockingScript(uint32(index), GetUnlockingScript(t, txDraft, uint32(index), privateKey))
			require.NoError(t, err)
		}

		// Record the final transaction
		var finalTx *Transaction
		finalTx, err = client.RecordTransaction(ctx, rawXPub, txDraft.String(),
			draftTransaction.ID,
			WithMetadata("custom_tag_2", "custom_value_2"),
		)
		require.NoError(t, err)
		require.NotNil(t, finalTx)

		// Check that the transaction was saved
		assert.Equal(t, draftTransaction.ID, finalTx.DraftID)
		assert.Equal(t, uint64(4903), finalTx.TotalValue)
		assert.Equal(t, uint64(97), finalTx.Fee)
	})
}
