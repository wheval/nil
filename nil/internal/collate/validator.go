package collate

import (
	"context"
	"fmt"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/network"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/rs/zerolog"
)

type Validator struct {
	params Params

	txFabric       db.DB
	pool           TxnPool
	networkManager *network.Manager

	logger zerolog.Logger
}

func (s *Validator) getBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	tx, err := s.txFabric.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	block, err := db.ReadBlock(tx, s.params.ShardId, hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (s *Validator) BuildProposal(ctx context.Context) (*execution.Proposal, error) {
	proposer := newProposer(s.params, s.params.Topology, s.pool, s.logger)
	proposal, err := proposer.GenerateProposal(ctx, s.txFabric)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proposal: %w", err)
	}
	return proposal, nil
}

func (s *Validator) VerifyProposal(ctx context.Context, proposal *execution.Proposal) (*types.Block, error) {
	prevBlock, err := s.getBlock(ctx, proposal.PrevBlockHash)
	if err != nil {
		return nil, err
	}

	gen, err := execution.NewBlockGenerator(ctx, s.params.BlockGeneratorParams, s.txFabric, prevBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to create block generator: %w", err)
	}
	defer gen.Rollback()

	gasPrices := gen.CollectGasPrices(proposal.PrevBlockId)
	res, err := gen.BuildBlock(proposal, gasPrices)
	if err != nil {
		return nil, fmt.Errorf("failed to generate block: %w", err)
	}
	return res.Block, nil
}

func (s *Validator) InsertProposal(ctx context.Context, proposal *execution.Proposal, params *types.ConsensusParams) error {
	prevBlock, err := s.getBlock(ctx, proposal.PrevBlockHash)
	if err != nil {
		return err
	}

	gen, err := execution.NewBlockGenerator(ctx, s.params.BlockGeneratorParams, s.txFabric, prevBlock)
	if err != nil {
		return fmt.Errorf("failed to create block generator: %w", err)
	}
	defer gen.Rollback()

	res, err := gen.GenerateBlock(proposal, params)
	if err != nil {
		return fmt.Errorf("failed to generate block: %w", err)
	}

	if err := s.pool.OnCommitted(ctx, proposal.ExternalTxns); err != nil {
		s.logger.Warn().Err(err).
			Msgf("Failed to remove %d committed transactions from pool", len(proposal.ExternalTxns))
	}

	return PublishBlock(ctx, s.networkManager, s.params.ShardId, &types.BlockWithExtractedData{
		Block:           res.Block,
		InTransactions:  res.InTxns,
		OutTransactions: res.OutTxns,
		ChildBlocks:     proposal.ShardHashes,
		Config:          res.ConfigParams,
	})
}
