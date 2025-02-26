package collate

import (
	"context"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/rollup"
	"github.com/NilFoundation/nil/nil/services/txnpool"
	"github.com/rs/zerolog"
)

type TxnPool interface {
	Peek(ctx context.Context, n int) ([]*types.TxnWithHash, error)
	Discard(ctx context.Context, txns []common.Hash, reason txnpool.DiscardReason) error
	OnCommitted(ctx context.Context, committed []*types.Transaction) error
}

type Consensus interface {
	RunSequence(ctx context.Context, height uint64) error
}

type Params struct {
	execution.BlockGeneratorParams

	MaxInternalGasInBlock          types.Gas
	MaxGasInBlock                  types.Gas
	MaxInternalTransactionsInBlock int
	MaxForwardTransactionsInBlock  int

	CollatorTickPeriod time.Duration
	Timeout            time.Duration

	Topology ShardTopology

	L1Fetcher rollup.L1BlockFetcher
}

type Scheduler struct {
	consensus      Consensus
	syncer         *Syncer
	txFabric       db.DB
	validator      *Validator
	networkManager *network.Manager

	params *Params

	logger zerolog.Logger

	l1Fetcher rollup.L1BlockFetcher
}

func NewScheduler(validator *Validator, txFabric db.DB, consensus Consensus, networkManager *network.Manager) *Scheduler {
	params := validator.params
	return &Scheduler{
		txFabric:       txFabric,
		validator:      validator,
		networkManager: networkManager,
		params:         params,
		logger: logging.NewLogger("collator").With().
			Stringer(logging.FieldShardId, params.ShardId).
			Logger(),
		l1Fetcher: params.L1Fetcher,
		consensus: consensus,
	}
}

func (s *Scheduler) Validator() *Validator {
	return s.validator
}

func (s *Scheduler) Run(ctx context.Context, syncer *Syncer, consensus Consensus) error {
	syncer.WaitComplete()

	s.logger.Info().Msg("Starting collation...")
	s.syncer = syncer

	// Enable handler for blocks relaying
	SetRequestHandler(ctx, s.networkManager, s.params.ShardId, s.txFabric, s.logger)

	ticker := time.NewTicker(s.params.CollatorTickPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.doCollate(ctx); err != nil {
				if ctx.Err() != nil {
					s.logger.Info().Msg("Stopping collation...")
					return nil
				}
				s.logger.Error().Err(err).Msg("Failed to collate")
			}
		case <-ctx.Done():
			s.logger.Info().Msg("Stopping collation...")
			return nil
		}
	}
}

func (s *Scheduler) doCollate(ctx context.Context) error {
	if s.params.DisableConsensus {
		proposal, err := s.validator.BuildProposal(ctx)
		if err != nil {
			return err
		}

		return s.validator.InsertProposal(ctx, proposal, &types.ConsensusParams{})
	} else {
		block, _, err := s.validator.GetLastBlock(ctx)
		if err != nil {
			return err
		}

		subId, syncCh := s.syncer.Subscribe()
		defer s.syncer.Unsubscribe(subId)

		ctx, cancelFn := context.WithCancel(ctx)
		defer cancelFn()

		consCh := make(chan error, 1)
		go func() {
			consCh <- s.consensus.RunSequence(ctx, block.Id.Uint64()+1)
		}()

		select {
		case <-syncCh:
			cancelFn()
			err := <-consCh
			s.logger.Debug().Err(err).Msg("Consensus interrupted by syncer")
			return nil
		case err := <-consCh:
			return err
		}
	}
}
