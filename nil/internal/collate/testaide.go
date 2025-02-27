//go:build test

package collate

import (
	"context"
	"slices"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/txnpool"
)

type MockTxnPool struct {
	Txns     []*types.Transaction
	MetaTxns []*types.TxnWithHash

	LastDiscarded []common.Hash
	LastReason    txnpool.DiscardReason
}

var _ TxnPool = (*MockTxnPool)(nil)

func (m *MockTxnPool) Reset() {
	m.Txns = m.Txns[:0]
	m.MetaTxns = m.MetaTxns[:0]
	m.LastDiscarded = nil
	m.LastReason = 0
}

func (m *MockTxnPool) Peek(n int) ([]*types.TxnWithHash, error) {
	if n > len(m.Txns) {
		return m.MetaTxns, nil
	}
	return m.MetaTxns[:n], nil
}

func (m *MockTxnPool) Discard(_ context.Context, txns []common.Hash, reason txnpool.DiscardReason) error {
	m.LastDiscarded = txns
	m.LastReason = reason
	return nil
}

func (m *MockTxnPool) OnCommitted(context.Context, types.Value, []*types.Transaction) error {
	return nil
}

func (m *MockTxnPool) Add(txns ...*types.Transaction) {
	m.Txns = append(m.Txns, txns...)

	m.MetaTxns = slices.Grow(m.MetaTxns, len(m.Txns))
	for _, txn := range txns {
		m.MetaTxns = append(m.MetaTxns, types.NewTxnWithHash(txn))
	}
}
