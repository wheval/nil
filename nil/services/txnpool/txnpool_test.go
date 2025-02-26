package txnpool

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
)

type SuiteTxnPool struct {
	suite.Suite

	ctx    context.Context
	cancel context.CancelFunc

	pool *TxnPool
}

func newTransaction(seqno types.Seqno, fee uint64) *types.Transaction {
	address := types.ShardAndHexToAddress(0, "deadbeef")
	return &types.Transaction{
		TransactionDigest: types.TransactionDigest{To: address, Seqno: seqno},
		Value:             types.NewValueFromUint64(fee),
	}
}

func (s *SuiteTxnPool) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	var err error
	s.pool, err = New(s.ctx, NewConfig(0), nil)
	s.Require().NoError(err)
}

func (s *SuiteTxnPool) TearDownTest() {
	s.cancel()
}

func (s *SuiteTxnPool) addTransactionsToPoolSuccessfully(pool Pool, txn ...*types.Transaction) {
	s.T().Helper()

	count := pool.TransactionCount()

	reasons, err := pool.Add(s.ctx, txn...)
	s.Require().NoError(err)
	s.Require().Len(reasons, len(txn))
	for _, reason := range reasons {
		s.Equal(NotSet, reason)
	}

	s.Equal(count+len(txn), pool.TransactionCount())
}

func (s *SuiteTxnPool) addTransactionsSuccessfully(txn ...*types.Transaction) {
	s.T().Helper()

	s.addTransactionsToPoolSuccessfully(s.pool, txn...)
}

func (s *SuiteTxnPool) addTransactionWithDiscardReason(txn *types.Transaction, reason DiscardReason) {
	s.T().Helper()

	count := s.pool.TransactionCount()

	reasons, err := s.pool.Add(s.ctx, txn)
	s.Require().NoError(err)
	s.Equal([]DiscardReason{reason}, reasons)

	s.Equal(count, s.pool.TransactionCount())
}

func (s *SuiteTxnPool) TestAdd() {
	wrongShardTxn := newTransaction(0, 123)
	wrongShardTxn.To = types.ShardAndHexToAddress(1, "deadbeef")
	_, err := s.pool.Add(s.ctx, wrongShardTxn)
	s.Require().Error(err)

	txn1 := newTransaction(0, 123)

	// Send the transaction for the first time - OK
	s.addTransactionsSuccessfully(txn1)

	// Send transaction once again - Duplicate hash
	s.addTransactionWithDiscardReason(txn1, DuplicateHash)

	// Send the same transaction with higher fee - OK
	// Doesn't use the same helper because here transaction count doesn't change
	txn2 := common.CopyPtr(txn1)
	txn2.FeeCredit = txn2.FeeCredit.Add64(1)
	reasons, err := s.pool.Add(s.ctx, txn2)
	s.Require().NoError(err)
	s.Require().Equal([]DiscardReason{NotSet}, reasons)
	s.Equal(1, s.pool.TransactionCount())

	// Send a different transaction with the same seqno - NotReplaced
	txn3 := common.CopyPtr(txn1)
	// Force the transaction to be different
	txn3.Data = append(txn3.Data, 0x01)
	s.Require().NotEqual(txn1.Hash(), txn3.Hash())
	// Add a higher fee (otherwise, no replacement can be expected anyway)
	txn3.FeeCredit = txn3.FeeCredit.Add64(1)
	s.addTransactionWithDiscardReason(txn3, NotReplaced)

	// Add a transaction with higher seqno to the same receiver
	s.addTransactionsSuccessfully(
		newTransaction(1, 124))

	// Add a transaction with lower seqno to the same receiver - SeqnoTooLow
	s.addTransactionWithDiscardReason(
		newTransaction(0, 124), SeqnoTooLow)

	// Add a transaction with higher seqno to a new receiver
	otherAddressTxn := newTransaction(1, 124)
	otherAddressTxn.To = types.ShardAndHexToAddress(0, "deadbeef01")
	s.addTransactionsSuccessfully(otherAddressTxn)
}

func (s *SuiteTxnPool) TestAddOverflow() {
	s.pool.cfg.Size = 1

	s.addTransactionsSuccessfully(
		newTransaction(0, 123))

	s.addTransactionWithDiscardReason(
		newTransaction(1, 123), PoolOverflow)
}

func (s *SuiteTxnPool) TestStarted() {
	s.True(s.pool.Started())
}

func (s *SuiteTxnPool) TestIdHashKnownGet() {
	txn := newTransaction(0, 123)
	s.addTransactionsSuccessfully(txn)

	has, err := s.pool.IdHashKnown(txn.Hash())
	s.Require().NoError(err)
	s.True(has)

	poolTxn, err := s.pool.Get(txn.Hash())
	s.Require().NoError(err)
	s.Equal(poolTxn, txn)

	has, err = s.pool.IdHashKnown(common.BytesToHash([]byte("abcd")))
	s.Require().NoError(err)
	s.False(has)

	poolTxn, err = s.pool.Get(common.BytesToHash([]byte("abcd")))
	s.Require().NoError(err)
	s.Nil(poolTxn)
}

func (s *SuiteTxnPool) TestSeqnoFromAddress() {
	txn := newTransaction(0, 123)

	_, inPool := s.pool.SeqnoToAddress(txn.To)
	s.Require().False(inPool)

	s.addTransactionsSuccessfully(txn)

	seqno, inPool := s.pool.SeqnoToAddress(txn.To)
	s.Require().True(inPool)
	s.Require().EqualValues(0, seqno)

	nextTxn := common.CopyPtr(txn)
	nextTxn.Seqno++
	s.addTransactionsSuccessfully(nextTxn)

	seqno, inPool = s.pool.SeqnoToAddress(txn.To)
	s.Require().True(inPool)
	s.Require().EqualValues(1, seqno)

	_, inPool = s.pool.SeqnoToAddress(types.BytesToAddress([]byte("abcd")))
	s.Require().False(inPool)
}

func (s *SuiteTxnPool) TestPeek() {
	address2 := types.ShardAndHexToAddress(0, "deadbeef02")

	txn21 := newTransaction(0, 123)
	txn21.To = address2
	txn22 := newTransaction(1, 123)
	txn22.To = address2

	s.addTransactionsSuccessfully(
		newTransaction(0, 123),
		newTransaction(1, 123),
		txn21,
		txn22)

	txns, err := s.pool.Peek(s.ctx, 1)
	s.Require().NoError(err)
	s.Len(txns, 1)

	txns, err = s.pool.Peek(s.ctx, 4)
	s.Require().NoError(err)
	s.Len(txns, 4)

	txns, err = s.pool.Peek(s.ctx, 10)
	s.Require().NoError(err)
	s.Len(txns, 4)
}

func (s *SuiteTxnPool) TestOnNewBlock() {
	address2 := types.ShardAndHexToAddress(0, "deadbeef02")

	txn11 := newTransaction(0, 123)
	txn12 := newTransaction(1, 123)

	txn21 := newTransaction(0, 123)
	txn21.To = address2
	txn22 := newTransaction(1, 123)
	txn22.To = address2

	s.addTransactionsSuccessfully(txn11, txn12, txn21, txn22)

	// TODO: Ideally we need to do that via execution state
	err := s.pool.OnCommitted(s.ctx, []*types.Transaction{txn11, txn12, txn21})
	s.Require().NoError(err)

	// After commit Peek should return only one transaction
	transactions, err := s.pool.Peek(s.ctx, 10)
	s.Require().NoError(err)
	s.Require().Len(transactions, 1)
	s.Equal(types.NewTxnWithHash(txn22), transactions[0])
	s.Equal(1, s.pool.TransactionCount())
}

func (s *SuiteTxnPool) TestNetwork() {
	nms := network.NewTestManagers(s.T(), s.ctx, 9100, 2)

	pool1, err := New(s.ctx, NewConfig(0), nms[0])
	s.Require().NoError(err)
	pool2, err := New(s.ctx, NewConfig(0), nms[1])
	s.Require().NoError(err)

	// Ensure that both nodes have subscribed, so that they will exchange this info on the following connect.
	s.Require().Eventually(func() bool {
		return slices.Contains(nms[0].PubSub().Topics(), topicPendingTransactions(0)) &&
			slices.Contains(nms[1].PubSub().Topics(), topicPendingTransactions(0))
	}, 1*time.Second, 50*time.Millisecond)

	network.ConnectManagers(s.T(), nms[0], nms[1])

	txn := newTransaction(0, 123)
	s.addTransactionsToPoolSuccessfully(pool1, txn)

	s.Eventually(func() bool {
		has, err := pool2.IdHashKnown(txn.Hash())
		s.Require().NoError(err)
		return has
	}, 20*time.Second, 200*time.Millisecond)
}

func TestSuiteTxnpool(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(SuiteTxnPool))
}

func BenchmarkTxnPoolAdd(b *testing.B) {
	shardId := types.ShardId(0)
	ctx := b.Context()
	pool, err := New(ctx, NewConfig(shardId), nil)
	if err != nil {
		b.Fatalf("Failed to create transaction pool: %s", err)
	}

	pool.cfg.Size = uint64(b.N)
	zerolog.SetGlobalLevel(zerolog.Disabled)

	txns := make([]*types.Transaction, b.N)
	var addr types.Address
	for i := range b.N {
		if i%2 == 0 {
			addr = types.GenerateRandomAddress(shardId)
		}
		txns[i] = &types.Transaction{
			TransactionDigest: types.TransactionDigest{Seqno: types.Seqno(i), To: addr},
		}
	}

	b.ResetTimer()

	for i := range b.N {
		_, err = pool.Add(ctx, txns[i])
		if err != nil {
			b.Fatalf("Failed to add transaction to pool: %s", err)
		}
	}
}
