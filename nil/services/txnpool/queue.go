package txnpool

type TxnQueue struct {
	data []*metaTxn
}

func NewTransactionQueue() *TxnQueue {
	return &TxnQueue{}
}

func (q *TxnQueue) Push(txn *metaTxn) {
	q.data = append(q.data, txn)
}

func (q *TxnQueue) Peek(n int) []*metaTxn {
	if len(q.data) < n {
		n = len(q.data)
	}
	return q.data[:n]
}

func (q *TxnQueue) Size() int {
	return len(q.data)
}

func (q *TxnQueue) Remove(txn *metaTxn) bool {
	for i, elem := range q.data {
		if elem.Seqno == txn.Seqno && elem.From.Equal(txn.From) {
			q.data = append(q.data[:i], q.data[i+1:]...)
			return true
		}
	}
	return false
}
