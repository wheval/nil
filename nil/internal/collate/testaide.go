//go:build test

package collate

import (
	"context"

	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/txnpool"
)

type MockTxnPool struct {
	Txns []*types.Transaction

	LastDiscarded []*types.Transaction
	LastReason    txnpool.DiscardReason
}

var _ TxnPool = (*MockTxnPool)(nil)

func (m *MockTxnPool) Reset() {
	m.Txns = nil
	m.LastDiscarded = nil
	m.LastReason = 0
}

func (m *MockTxnPool) Peek(_ context.Context, n int) ([]*types.Transaction, error) {
	if n > len(m.Txns) {
		return m.Txns, nil
	}
	return m.Txns[:n], nil
}

func (m *MockTxnPool) Discard(_ context.Context, txns []*types.Transaction, reason txnpool.DiscardReason) error {
	m.LastDiscarded = txns
	m.LastReason = reason
	return nil
}

func (m *MockTxnPool) OnCommitted(context.Context, []*types.Transaction) error {
	return nil
}
