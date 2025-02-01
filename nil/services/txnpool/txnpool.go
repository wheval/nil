package txnpool

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/check"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/rs/zerolog"
)

type Pool interface {
	Add(ctx context.Context, txns ...*types.Transaction) ([]DiscardReason, error)
	Discard(ctx context.Context, txns []*types.Transaction, reason DiscardReason) error
	OnCommitted(ctx context.Context, committed []*types.Transaction) error
	// IdHashKnown check whether transaction with given Id hash is known to the pool
	IdHashKnown(hash common.Hash) (bool, error)
	Started() bool

	Peek(ctx context.Context, n int) ([]*types.Transaction, error)
	SeqnoToAddress(addr types.Address) (seqno types.Seqno, inPool bool)
	TransactionCount() int
	Get(hash common.Hash) (*types.Transaction, error)
}

type metaTxn struct {
	*types.Transaction
	hash common.Hash
}

func newMetaTxn(txn *types.Transaction) *metaTxn {
	return &metaTxn{
		Transaction: txn,
		hash:        txn.Hash(),
	}
}

type TxnPool struct {
	started bool
	cfg     Config

	networkManager *network.Manager

	lock sync.Mutex

	byHash map[string]*metaTxn // hash => txn : only those records not committed to db yet
	all    *ByReceiverAndSeqno // from => (sorted map of txn seqno => *txn)
	queue  *TxnQueue
	logger zerolog.Logger
}

func New(ctx context.Context, cfg Config, networkManager *network.Manager) (*TxnPool, error) {
	logger := logging.NewLogger("txnpool").With().
		Stringer(logging.FieldShardId, cfg.ShardId).
		Logger()

	res := &TxnPool{
		started: true,
		cfg:     cfg,

		networkManager: networkManager,

		byHash: map[string]*metaTxn{},
		all:    NewBySenderAndSeqno(logger),
		queue:  NewTransactionQueue(),
		logger: logger,
	}

	if networkManager == nil {
		// we don't always want to run the network (e.g., in tests)
		return res, nil
	}

	sub, err := networkManager.PubSub().Subscribe(topicPendingTransactions(cfg.ShardId))
	if err != nil {
		return nil, err
	}

	go func() {
		res.listen(ctx, sub)
	}()

	return res, nil
}

func (p *TxnPool) listen(ctx context.Context, sub *network.Subscription) {
	defer sub.Close()

	for m := range sub.Start(ctx, true) {
		txn := &types.Transaction{}
		if err := txn.UnmarshalSSZ(m); err != nil {
			p.logger.Error().Err(err).
				Msg("Failed to unmarshal transaction from network")
			continue
		}

		mm := newMetaTxn(txn)
		reasons, err := p.add(mm)
		if err != nil {
			p.logger.Error().Err(err).
				Stringer(logging.FieldTransactionHash, mm.hash).
				Msg("Failed to add transaction from network")
			continue
		}

		if reasons[0] != NotSet {
			p.logger.Debug().
				Stringer(logging.FieldTransactionHash, mm.hash).
				Msgf("Discarded transaction from network with reason %s", reasons[0])
		}
	}
}

func (p *TxnPool) Add(ctx context.Context, txns ...*types.Transaction) ([]DiscardReason, error) {
	mms := make([]*metaTxn, len(txns))
	for i, txn := range txns {
		mms[i] = newMetaTxn(txn)
	}

	reasons, err := p.add(mms...)
	if err != nil {
		return nil, err
	}

	for i, mm := range mms {
		if reasons[i] != NotSet {
			continue
		}

		if err := PublishPendingTransaction(ctx, p.networkManager, p.cfg.ShardId, mm); err != nil {
			p.logger.Error().Err(err).
				Stringer(logging.FieldTransactionHash, mm.hash).
				Msg("Failed to publish transaction to network")
		}
	}

	return reasons, nil
}

func (p *TxnPool) add(txns ...*metaTxn) ([]DiscardReason, error) {
	discardReasons := make([]DiscardReason, len(txns))

	p.lock.Lock()
	defer p.lock.Unlock()

	for i, txn := range txns {
		if txn.To.ShardId() != p.cfg.ShardId {
			return nil, fmt.Errorf("transaction shard id %d does not match pool shard id %d", txn.To.ShardId(), p.cfg.ShardId)
		}

		if reason, ok := p.validateTxn(txn); !ok {
			discardReasons[i] = reason
			continue
		}

		if _, ok := p.byHash[string(txn.hash.Bytes())]; ok {
			discardReasons[i] = DuplicateHash
			continue
		}

		if reason := p.addLocked(txn); reason != NotSet {
			discardReasons[i] = reason
			continue
		}
		discardReasons[i] = NotSet // unnecessary
		p.logger.Debug().
			Uint64(logging.FieldShardId, uint64(txn.To.ShardId())).
			Stringer(logging.FieldTransactionHash, txn.hash).
			Stringer(logging.FieldTransactionTo, txn.To).
			Msg("Added new transaction.")
	}

	return discardReasons, nil
}

func (p *TxnPool) validateTxn(txn *metaTxn) (DiscardReason, bool) {
	seqno, has := p.all.seqno(txn.To)
	if has && seqno > txn.Seqno {
		p.logger.Debug().
			Uint64(logging.FieldShardId, uint64(txn.To.ShardId())).
			Stringer(logging.FieldTransactionHash, txn.hash).
			Uint64(logging.FieldAccountSeqno, seqno.Uint64()).
			Uint64(logging.FieldTransactionSeqno, txn.Seqno.Uint64()).
			Msg("Seqno too low.")
		return SeqnoTooLow, false
	}

	if txn.ChainId != types.DefaultChainId {
		return InvalidChainId, false
	}

	return NotSet, true
}

func (p *TxnPool) idHashKnownLocked(hash common.Hash) bool {
	if _, ok := p.byHash[string(hash.Bytes())]; ok {
		return true
	}
	return false
}

func (p *TxnPool) IdHashKnown(hash common.Hash) (bool, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.idHashKnownLocked(hash), nil
}

func (p *TxnPool) Started() bool {
	return p.started
}

func (p *TxnPool) SeqnoToAddress(addr types.Address) (seqno types.Seqno, inPool bool) {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.all.seqno(addr)
}

func (p *TxnPool) Get(hash common.Hash) (*types.Transaction, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	txn := p.getLocked(hash)
	if txn == nil {
		return nil, nil
	}
	return txn.Transaction, nil
}

func (p *TxnPool) getLocked(hash common.Hash) *metaTxn {
	txn, ok := p.byHash[string(hash.Bytes())]
	if ok {
		return txn
	}
	return nil
}

func shouldReplace(existing, candidate *metaTxn) bool {
	if candidate.FeeCredit.Cmp(existing.FeeCredit) <= 0 {
		return false
	}

	// Discard the previous transaction if it is the same but at a lower fee
	existingFee := existing.FeeCredit
	existing.FeeCredit = candidate.FeeCredit
	defer func() {
		existing.FeeCredit = existingFee
	}()

	return bytes.Equal(existing.Hash().Bytes(), candidate.hash.Bytes())
}

func (p *TxnPool) addLocked(txn *metaTxn) DiscardReason {
	// Insert to pending pool, if pool doesn't have a txn with the same dst and seqno.
	// If pool has a txn with the same dst and seqno, only fee bump is possible; otherwise NotReplaced is returned.
	found := p.all.get(txn.To, txn.Seqno)
	if found != nil {
		if !shouldReplace(found, txn) {
			return NotReplaced
		}

		p.queue.Remove(found)
		p.discardLocked(found, ReplacedByHigherTip)
	}

	if uint64(p.queue.Size()) >= p.cfg.Size {
		return PoolOverflow
	}

	hashStr := string(txn.hash.Bytes())
	p.byHash[hashStr] = txn

	replaced := p.all.replaceOrInsert(txn)
	check.PanicIfNot(replaced == nil)

	p.queue.Push(txn)
	return NotSet
}

// dropping transaction from all sub-structures and from db
// Important: don't call it while iterating by "all"
func (p *TxnPool) discardLocked(mm *metaTxn, reason DiscardReason) {
	hashStr := string(mm.hash.Bytes())
	delete(p.byHash, hashStr)
	p.all.delete(mm, reason)
}

func (p *TxnPool) Discard(_ context.Context, txns []*types.Transaction, reason DiscardReason) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, txn := range txns {
		mm := p.getLocked(txn.Hash())
		if mm == nil {
			continue
		}

		p.queue.Remove(mm)
		p.discardLocked(mm, reason)
	}

	return nil
}

func (p *TxnPool) OnCommitted(_ context.Context, committed []*types.Transaction) (err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.removeCommitted(p.all, committed)
}

// removeCommitted - apply new highest block (or batch of blocks)
//
// 1. New block arrives, which potentially changes the balance and the seqno of some senders.
// We use senderIds data structure to find relevant senderId values, and then use senders data structure to
// modify state_balance and state_seqno, potentially remove some elements (if transaction with some seqno is
// included into a block), and finally, walk over the transaction records and update queue depending on
// the actual presence of seqno gaps and what the balance is.
func (p *TxnPool) removeCommitted(bySeqno *ByReceiverAndSeqno, txns []*types.Transaction) error { //nolint:unparam
	seqnosToRemove := map[types.Address]types.Seqno{}
	for _, txn := range txns {
		seqno, ok := seqnosToRemove[txn.To]
		if !ok || txn.Seqno > seqno {
			seqnosToRemove[txn.To] = txn.Seqno
		}
	}

	var toDel []*metaTxn // can't delete items while iterate them

	discarded := 0

	for senderID, seqno := range seqnosToRemove {
		bySeqno.ascend(senderID, func(txn *metaTxn) bool {
			if txn.Seqno > seqno {
				p.logger.Trace().
					Uint64(logging.FieldShardId, uint64(txn.To.ShardId())).
					Uint64(logging.FieldTransactionSeqno, txn.Seqno.Uint64()).
					Uint64(logging.FieldAccountSeqno, seqno.Uint64()).
					Msg("Removing committed, cmp seqnos")

				return false
			}

			p.logger.Trace().
				Uint64(logging.FieldShardId, uint64(txn.To.ShardId())).
				Stringer(logging.FieldTransactionHash, txn.hash).
				Stringer(logging.FieldTransactionTo, txn.To).
				Uint64(logging.FieldTransactionSeqno, txn.Seqno.Uint64()).
				Msg("Remove committed.")

			toDel = append(toDel, txn)
			p.queue.Remove(txn)
			return true
		})

		discarded += len(toDel)

		for _, txn := range toDel {
			p.discardLocked(txn, Committed)
		}
		toDel = toDel[:0]
	}

	if discarded > 0 {
		p.logger.Debug().
			Int("count", discarded).
			Msg("Discarded transactions")
	}

	return nil
}

func (p *TxnPool) Peek(ctx context.Context, n int) ([]*types.Transaction, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	mms := p.queue.Peek(n)
	txns := make([]*types.Transaction, len(mms))
	for i, mm := range mms {
		txns[i] = mm.Transaction
	}
	return txns, nil
}

func (p *TxnPool) TransactionCount() int {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.queue.Size()
}
