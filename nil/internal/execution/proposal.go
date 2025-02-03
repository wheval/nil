package execution

import (
	"slices"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
)

type Proposal struct {
	PrevBlockId   types.BlockNumber   `json:"prevBlockId"`
	PrevBlockHash common.Hash         `json:"prevBlockHash"`
	CollatorState types.CollatorState `json:"collatorState"`
	MainChainHash common.Hash         `json:"mainChainHash"`
	ShardHashes   []common.Hash       `json:"shardHashes" ssz-max:"4096"`

	InternalTxns []*types.Transaction `json:"internalTxns" ssz-max:"4096"`
	ExternalTxns []*types.Transaction `json:"externalTxns" ssz-max:"4096"`
	ForwardTxns  []*types.Transaction `json:"forwardTxns" ssz-max:"4096"`
}

func NewEmptyProposal() *Proposal {
	return &Proposal{}
}

func (p *Proposal) GetMainShardHash(shardId types.ShardId) *common.Hash {
	if shardId.IsMainShard() {
		return &p.PrevBlockHash
	}
	return &p.MainChainHash
}

func SplitTransactions(transactions []*types.Transaction, f func(t *types.Transaction) bool) (a, b []*types.Transaction) {
	if pos := slices.IndexFunc(transactions, f); pos != -1 {
		return transactions[:pos], transactions[pos:]
	}

	return transactions, nil
}

// SplitInTransactions splits incoming transactions in the block into internal and external ones.
// Internal transactions come before the external ones.
func SplitInTransactions(transactions []*types.Transaction) (internal, external []*types.Transaction) {
	return SplitTransactions(transactions, func(t *types.Transaction) bool {
		return t.IsExternal()
	})
}

// SplitOutTransactions splits outgoing transactions in the block into forwarded and generated ones.
// Forwarded transactions come before the generated ones.
func SplitOutTransactions(transactions []*types.Transaction, shardId types.ShardId) (forwarded, generated []*types.Transaction) {
	return SplitTransactions(transactions, func(t *types.Transaction) bool {
		return t.From.ShardId() == shardId
	})
}
