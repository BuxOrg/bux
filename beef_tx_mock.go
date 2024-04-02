package bux

import (
	"context"
)

// MockTransactionStore is a mock implementation of the TransactionStore interface
type MockTransactionStore struct {
	Transactions map[string]*Transaction
}

// NewMockTransactionStore creates a new instance of the MockTransactionStore
func NewMockTransactionStore() *MockTransactionStore {
	return &MockTransactionStore{
		Transactions: make(map[string]*Transaction),
	}
}

// AddToStore adds a transaction to the store
func (m *MockTransactionStore) AddToStore(tx *Transaction) {
	m.Transactions[tx.ID] = tx
}

// GetTransactionsByIDs returns transactions by their IDs
func (m *MockTransactionStore) GetTransactionsByIDs(ctx context.Context, txIDs []string) ([]*Transaction, error) {
	var txs []*Transaction
	for _, txID := range txIDs {
		if tx, exists := m.Transactions[txID]; exists {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}
