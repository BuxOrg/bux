package bux

import (
	"context"
	"fmt"
)

type MockTransactionGetter struct {
	Transactions map[string]*Transaction
}

func NewMockTransactionGetter() *MockTransactionGetter {
	return &MockTransactionGetter{
		Transactions: make(map[string]*Transaction),
	}
}

func (m *MockTransactionGetter) Init(transactions []*Transaction) {
	for _, tx := range transactions {
		m.Transactions[tx.ID] = tx
	}
}

func (m *MockTransactionGetter) GetTransactionByID(ctx context.Context, txID string) (*Transaction, error) {
	if tx, exists := m.Transactions[txID]; exists {
		return tx, nil
	}
	return nil, fmt.Errorf("no records found")
}
