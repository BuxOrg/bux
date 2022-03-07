package bux

import (
	"testing"

	"github.com/BuxOrg/bux/utils"
	"github.com/stretchr/testify/assert"
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
		assert.Len(t, AllModelNames, 8)
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
