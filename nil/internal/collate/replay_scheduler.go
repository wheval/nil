package collate

import (
	"context"
	"errors"
	"slices"
	"sort"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/rs/zerolog"
)

type ReplayParams struct {
	execution.BlockGeneratorParams

	Timeout time.Duration

	ReplayFirstBlock types.BlockNumber
	ReplayLastBlock  types.BlockNumber
}

type ReplayScheduler struct {
	txFabric db.DB

	params ReplayParams

	logger zerolog.Logger
}

func NewReplayScheduler(txFabric db.DB, params ReplayParams) *ReplayScheduler {
	return &ReplayScheduler{
		txFabric: txFabric,
		params:   params,
		logger: logging.NewLogger("block-replayer").With().
			Stringer(logging.FieldShardId, params.ShardId).
			Logger(),
	}
}

func (s *ReplayScheduler) Run(ctx context.Context) error {
	s.logger.Info().Msgf("Starting block replay for blocks [%d - %d]...", s.params.ReplayFirstBlock, s.params.ReplayLastBlock)

runloop:
	for blockId := s.params.ReplayFirstBlock; blockId <= s.params.ReplayLastBlock; blockId++ {
		if ctx.Err() != nil {
			break runloop
		}

		if err := s.doReplay(ctx, blockId); err != nil {
			return err
		}
	}

	<-ctx.Done()
	s.logger.Info().Msg("Stopping block replay...")
	if !errors.Is(ctx.Err(), context.Canceled) {
		return ctx.Err()
	}
	return nil
}

func (s *ReplayScheduler) doReplay(ctx context.Context, blockId types.BlockNumber) error {
	ctx, cancel := context.WithTimeout(ctx, s.params.Timeout)
	defer cancel()

	proposal, err := s.buildProposalFromPrevBlock(ctx, blockId)
	if err != nil {
		return err
	}

	gen, err := execution.NewBlockGenerator(ctx, s.params.BlockGeneratorParams, s.txFabric, proposal.GetMainShardHash(s.params.ShardId))
	if err != nil {
		return err
	}
	defer gen.Rollback()

	if _, err := gen.GenerateBlock(proposal, s.logger, nil); err != nil {
		return err
	}

	return nil
}

func (s *ReplayScheduler) switchLastBlock(ctx context.Context, blockId types.BlockNumber) (common.Hash, error) {
	rwTx, err := s.txFabric.CreateRwTx(ctx)
	if err != nil {
		return common.EmptyHash, err
	}
	defer rwTx.Rollback()

	blockHash, err := db.ReadBlockHashByNumber(rwTx, s.params.ShardId, blockId)
	if err != nil {
		return common.EmptyHash, err
	}

	s.logger.Debug().Msgf("Switching last block to %s", blockHash)
	if err = db.WriteLastBlockHash(rwTx, s.params.ShardId, blockHash); err != nil {
		return common.EmptyHash, err
	}

	return blockHash, rwTx.Commit()
}

func (s *ReplayScheduler) buildProposalFromPrevBlock(ctx context.Context, blockId types.BlockNumber) (*execution.Proposal, error) {
	if s.params.ShardId == types.MainShardId {
		return nil, errors.New("replay for masterchain is not supported")
	}
	if blockId == types.BlockNumber(0) {
		return nil, errors.New("replay for zerostate-block is not supported")
	}

	proposal := &execution.Proposal{PrevBlockId: blockId - 1}

	// NOTE: masterchain last block isn't switched now
	if hash, err := s.switchLastBlock(ctx, proposal.PrevBlockId); err != nil {
		return nil, err
	} else {
		proposal.PrevBlockHash = hash
	}

	roTx, err := s.txFabric.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer roTx.Rollback()

	prevBlock, err := db.ReadBlock(roTx, s.params.ShardId, proposal.PrevBlockHash)
	if err != nil {
		return nil, err
	}
	blockToReplay, err := db.ReadBlockByNumber(roTx, s.params.ShardId, blockId)
	if err != nil {
		return nil, err
	}

	proposal.MainChainHash = prevBlock.MainChainHash
	s.logger.Trace().Msgf("Last block is %s, last MC block is %s", proposal.PrevBlockHash, proposal.MainChainHash)

	// we could also consider option with fairly collecting these transactions
	// from neighbor shards and running proposer
	// however it's not a purpose of replay mode (at least now)
	inTxns, err := s.collectTxns(roTx, blockToReplay.InTransactionsRoot)
	if err != nil {
		return nil, err
	}
	proposal.InternalTxns, proposal.ExternalTxns = execution.SplitInTransactions(inTxns)

	forwardTxns, err := s.collectTxns(roTx, blockToReplay.OutTransactionsRoot)
	if err != nil {
		return nil, err
	}
	proposal.ForwardTxns, _ = execution.SplitOutTransactions(forwardTxns, s.params.ShardId)

	return proposal, nil
}

func (s *ReplayScheduler) collectTxns(roTx db.RoTx, root common.Hash) ([]*types.Transaction, error) {
	inTxnsReader := execution.NewDbTransactionTrieReader(roTx, s.params.ShardId)
	inTxnsReader.SetRootHash(root)
	entries, err := inTxnsReader.Entries()
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })
	return slices.Collect(common.Transform(slices.Values(entries),
		func(e execution.Entry[types.TransactionIndex, *types.Transaction]) *types.Transaction {
			return e.Val
		})), nil
}
