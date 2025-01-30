package collate

import (
	"context"
	"time"

	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/txnpool"
	"github.com/rs/zerolog"
)

type TxnPool interface {
	Peek(ctx context.Context, n int) ([]*types.Transaction, error)
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

	ZeroState       string
	ZeroStateConfig *execution.ZeroStateConfig

	Topology ShardTopology
}

type Scheduler struct {
	consensus      Consensus
	txFabric       db.DB
	pool           txnpool.Pool
	networkManager *network.Manager

	params Params

	logger zerolog.Logger
}

func NewScheduler(txFabric db.DB, pool txnpool.Pool, params Params, networkManager *network.Manager) *Scheduler {
	return &Scheduler{
		txFabric:       txFabric,
		pool:           pool,
		networkManager: networkManager,
		params:         params,
		logger: logging.NewLogger("collator").With().
			Stringer(logging.FieldShardId, params.ShardId).
			Logger(),
	}
}

func (s *Scheduler) Validator() *Validator {
	return &Validator{
		params:         s.params,
		txFabric:       s.txFabric,
		pool:           s.pool,
		networkManager: s.networkManager,
		logger:         s.logger,
	}
}

func (s *Scheduler) Run(ctx context.Context, syncer *Syncer, consensus Consensus) error {
	syncer.WaitComplete()

	s.logger.Info().Msg("Starting collation...")
	s.consensus = consensus

	// Enable handler for snapshot relaying
	SetBootstrapHandler(ctx, s.networkManager, s.params.ShardId, s.txFabric)

	// Enable handler for blocks relaying
	SetRequestHandler(ctx, s.networkManager, s.params.ShardId, s.txFabric)

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
	id, err := s.readLastBlockId(ctx)
	if err != nil {
		return err
	}

	return s.consensus.RunSequence(ctx, id.Uint64()+1)
}

func (s *Scheduler) readLastBlockId(ctx context.Context) (types.BlockNumber, error) {
	roTx, err := s.txFabric.CreateRoTx(ctx)
	if err != nil {
		return 0, err
	}
	defer roTx.Rollback()

	b, _, err := db.ReadLastBlock(roTx, s.params.ShardId)
	if err != nil {
		return 0, err
	}

	return b.Id, nil
}
