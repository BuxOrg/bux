package bux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSyncTransaction_GetModelName will test the method GetModelName()
func TestSyncTransaction_GetModelName(t *testing.T) {
	t.Parallel()

	bTx := newSyncTransaction(testTxID, nil, New())
	assert.Equal(t, ModelSyncTransaction.String(), bTx.GetModelName())
}
