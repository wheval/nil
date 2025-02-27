package txnpool

import (
	"container/heap"

	"github.com/NilFoundation/nil/nil/common/check"
)

type TxnQueue struct {
	txns []*metaTxn
}

func (p *TxnQueue) Clone() *TxnQueue {
	txns := make([]*metaTxn, len(p.txns))
	for i, txn := range p.txns {
		txns[i] = new(metaTxn)
		*txns[i] = *txn
	}
	return &TxnQueue{txns: txns}
}

func (p *TxnQueue) Len() int {
	return len(p.txns)
}

func (p *TxnQueue) Less(i, j int) bool {
	return p.txns[i].effectivePriorityFee.Cmp(p.txns[j].effectivePriorityFee) > 0
}

func (p *TxnQueue) Swap(i, j int) {
	p.txns[i], p.txns[j] = p.txns[j], p.txns[i]
	p.txns[i].bestIndex = i
	p.txns[j].bestIndex = j
}

func (p *TxnQueue) Push(x any) {
	txn, ok := x.(*metaTxn)
	check.PanicIfNot(ok)
	txn.bestIndex = len(p.txns)
	p.txns = append(p.txns, txn)
}

func (p *TxnQueue) Pop() any {
	old := p.txns
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	p.txns = old[0 : n-1]
	return item
}

func (p *TxnQueue) Remove(txn *metaTxn) {
	if txn.bestIndex >= 0 {
		heap.Remove(p, txn.bestIndex)
		txn.bestIndex = -1
	}
}
