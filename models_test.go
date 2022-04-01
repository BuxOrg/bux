package bux

import (
	"testing"
	"time"

	"github.com/BuxOrg/bux/utils"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMetadataKey   = "test_key"
	testMetadataValue = "test_value"
)

// TestModelName_String will test the method String()
func TestModelName_String(t *testing.T) {
	t.Parallel()

	t.Run("all model names", func(t *testing.T) {
		assert.Equal(t, "destination", ModelDestination.String())
		assert.Equal(t, "empty", ModelNameEmpty.String())
		assert.Equal(t, "incoming_transaction", ModelIncomingTransaction.String())
		assert.Equal(t, "metadata", ModelMetadata.String())
		assert.Equal(t, "sync_transaction", ModelSyncTransaction.String())
		assert.Equal(t, "transaction", ModelTransaction.String())
		assert.Equal(t, "utxo", ModelUtxo.String())
		assert.Equal(t, "xpub", ModelXPub.String())
		assert.Equal(t, "paymail", ModelPaymail.String())
		assert.Equal(t, "block_header", ModelBlockHeader.String())
		assert.Len(t, AllModelNames, 11)
	})
}

// TestModelName_IsEmpty will test the method IsEmpty()
func TestModelName_IsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("empty model", func(t *testing.T) {
		assert.Equal(t, true, ModelNameEmpty.IsEmpty())
		assert.Equal(t, false, ModelUtxo.IsEmpty())
	})
}

// TestModel_GetModelName will test the GetModelName function
func TestModel_GetModelName(t *testing.T) {
	t.Parallel()

	t.Run("models", func(t *testing.T) {
		assert.Nil(t, utils.GetModelName(nil))
		transaction := Transaction{}
		assert.Equal(t, ModelTransaction.String(), *utils.GetModelName(transaction))
		xPub := Xpub{}
		assert.Equal(t, ModelXPub.String(), *utils.GetModelName(xPub))
		destination := Destination{}
		assert.Equal(t, ModelDestination.String(), *utils.GetModelName(destination))
		utxo := Utxo{}
		assert.Equal(t, ModelUtxo.String(), *utils.GetModelName(utxo))
	})
}

// TestModel_GetModelTableName will test the GetModelTableName function
func TestModel_GetModelTableName(t *testing.T) {
	t.Parallel()

	t.Run("models", func(t *testing.T) {
		assert.Nil(t, utils.GetModelName(nil))
		transaction := Transaction{}
		assert.Equal(t, tableTransactions, *utils.GetModelTableName(transaction))
		xPub := Xpub{}
		assert.Equal(t, tableXPubs, *utils.GetModelTableName(xPub))
		destination := Destination{}
		assert.Equal(t, tableDestinations, *utils.GetModelTableName(destination))
		utxo := Utxo{}
		assert.Equal(t, tableUTXOs, *utils.GetModelTableName(utxo))
	})
}

func (ts *EmbeddedDBTestSuite) createXpubModels(tc *TestingClient, t *testing.T, number int) *TestingClient {
	for i := 0; i < number; i++ {
		_, xPublicKey, err := bitcoin.GenerateHDKeyPair(bitcoin.SecureSeedLength)
		require.NoError(t, err)
		xPub := newXpub(xPublicKey, append(tc.client.DefaultModelOptions(), New())...)
		xPub.CurrentBalance = 125000
		xPub.NextExternalNum = 12
		xPub.NextInternalNum = 37
		err = xPub.Save(tc.ctx)
		require.NoError(t, err)
	}

	return tc
}

type xPubFieldsTest struct {
	CurrentBalance uint64 `json:"current_balance" toml:"current_balance" yaml:"current_balance" bson:"current_balance"`
}

// TestModels_GetModels will test the method GetModels()
func (ts *EmbeddedDBTestSuite) TestModels_GetModels() {

	numberOfModels := 10
	for _, testCase := range dbTestCases {
		ts.T().Run(testCase.name+" - GetModels", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)
			ts.createXpubModels(tc, t, numberOfModels)

			var models []*Xpub
			err := tc.client.Datastore().GetModels(
				tc.ctx,
				&models,
				nil,
				10,
				0,
				"",
				"",
				nil,
				30*time.Second,
			)
			require.NoError(t, err)
			require.Len(t, models, numberOfModels)
			assert.Equal(t, uint64(125000), models[0].CurrentBalance) // should be set
			assert.Equal(t, uint32(12), models[0].NextExternalNum)    // should be set
			assert.Equal(t, uint32(37), models[0].NextInternalNum)    // should be set
		})

		ts.T().Run(testCase.name+" - GetModels with projection", func(t *testing.T) {
			tc := ts.genericDBClient(t, testCase.database, false)
			defer tc.Close(tc.ctx)
			ts.createXpubModels(tc, t, numberOfModels)

			var models []*Xpub
			var results []*xPubFieldsTest
			err := tc.client.Datastore().GetModels(
				tc.ctx,
				&models,
				nil,
				10,
				0,
				"",
				"",
				&results,
				30*time.Second,
			)
			require.NoError(t, err)
			require.Len(t, results, numberOfModels)
			assert.Equal(t, uint64(125000), results[0].CurrentBalance) // should be set
		})
	}
}
