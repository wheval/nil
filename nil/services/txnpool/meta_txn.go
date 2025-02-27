package txnpool

import (
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/types"
)

type metaTxn struct {
	*types.TxnWithHash
	effectivePriorityFee types.Value
	bestIndex            int
	valid                bool
}

func newMetaTxn(txn *types.Transaction, baseFee types.Value) *metaTxn {
	effectivePriorityFee, valid := execution.GetEffectivePriorityFee(baseFee, txn)
	return &metaTxn{
		TxnWithHash:          types.NewTxnWithHash(txn),
		effectivePriorityFee: effectivePriorityFee,
		valid:                valid,
		bestIndex:            -1,
	}
}

func (m *metaTxn) IsValid() bool {
	return m.valid
}

func (m *metaTxn) IsInQueue() bool {
	return m.bestIndex >= 0
}
