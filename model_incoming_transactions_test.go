package bux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIncomingTransaction_GetModelName will test the method GetModelName()
func TestIncomingTransaction_GetModelName(t *testing.T) {
	t.Parallel()

	bTx := newIncomingTransaction(testTxID, testTxHex, New())
	assert.Equal(t, ModelIncomingTransaction.String(), bTx.GetModelName())
}
