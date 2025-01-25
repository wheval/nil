//go:build test

package collate

import (
	"context"

	"github.com/NilFoundation/nil/nil/internal/types"
)

type MockTxnPool struct {
	Txns []*types.Transaction
}

var _ TxnPool = (*MockTxnPool)(nil)

func (m *MockTxnPool) Peek(_ context.Context, n int) ([]*types.Transaction, error) {
	if n > len(m.Txns) {
		return m.Txns, nil
	}
	return m.Txns[:n], nil
}

func (m *MockTxnPool) OnCommitted(context.Context, []*types.Transaction) error {
	return nil
}
