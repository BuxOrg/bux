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
		assert.Equal(t, "paymail_address", ModelPaymailAddress.String())
		assert.Equal(t, "sync_transaction", ModelSyncTransaction.String())
		assert.Equal(t, "transaction", ModelTransaction.String())
		assert.Equal(t, "utxo", ModelUtxo.String())
		assert.Equal(t, "xpub", ModelXPub.String())
		assert.Len(t, AllModelNames, 9)
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

	t.Run("empty model", func(t *testing.T) {
		assert.Nil(t, utils.GetModelName(nil))
	})

	t.Run("base model names", func(t *testing.T) {

		xPub := Xpub{}
		assert.Equal(t, ModelXPub.String(), *utils.GetModelName(xPub))

		destination := Destination{}
		assert.Equal(t, ModelDestination.String(), *utils.GetModelName(destination))

		utxo := Utxo{}
		assert.Equal(t, ModelUtxo.String(), *utils.GetModelName(utxo))

		transaction := Transaction{}
		assert.Equal(t, ModelTransaction.String(), *utils.GetModelName(transaction))

		accessKey := AccessKey{}
		assert.Equal(t, ModelAccessKey.String(), *utils.GetModelName(accessKey))

		draftTx := DraftTransaction{}
		assert.Equal(t, ModelDraftTransaction.String(), *utils.GetModelName(draftTx))

		incomingTx := IncomingTransaction{}
		assert.Equal(t, ModelIncomingTransaction.String(), *utils.GetModelName(incomingTx))

		paymailAddress := PaymailAddress{}
		assert.Equal(t, ModelPaymailAddress.String(), *utils.GetModelName(paymailAddress))

		syncTx := SyncTransaction{}
		assert.Equal(t, ModelSyncTransaction.String(), *utils.GetModelName(syncTx))
	})
}

// TestModel_GetModelTableName will test the GetModelTableName function
func TestModel_GetModelTableName(t *testing.T) {
	t.Parallel()

	t.Run("empty model", func(t *testing.T) {
		assert.Nil(t, utils.GetModelTableName(nil))
	})

	t.Run("get model table names", func(t *testing.T) {
		xPub := Xpub{}
		assert.Equal(t, tableXPubs, *utils.GetModelTableName(xPub))

		destination := Destination{}
		assert.Equal(t, tableDestinations, *utils.GetModelTableName(destination))

		utxo := Utxo{}
		assert.Equal(t, tableUTXOs, *utils.GetModelTableName(utxo))

		transaction := Transaction{}
		assert.Equal(t, tableTransactions, *utils.GetModelTableName(transaction))

		accessKey := AccessKey{}
		assert.Equal(t, tableAccessKeys, *utils.GetModelTableName(accessKey))

		draftTx := DraftTransaction{}
		assert.Equal(t, tableDraftTransactions, *utils.GetModelTableName(draftTx))

		incomingTx := IncomingTransaction{}
		assert.Equal(t, tableIncomingTransactions, *utils.GetModelTableName(incomingTx))

		paymailAddress := PaymailAddress{}
		assert.Equal(t, tablePaymailAddresses, *utils.GetModelTableName(paymailAddress))

		syncTx := SyncTransaction{}
		assert.Equal(t, tableSyncTransactions, *utils.GetModelTableName(syncTx))
	})
}
