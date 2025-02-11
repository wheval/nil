package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/concurrent"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/metrics"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/rollupcontract"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/storage"
	scTypes "github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/rs/zerolog"
)

type Proposer struct {
	blockStorage storage.BlockStorage
	retryRunner  common.RetryRunner

	rollupContract *rollupcontract.Wrapper
	params         *ProposerParams

	metrics ProposerMetrics
	logger  zerolog.Logger
}

type ProposerParams struct {
	Endpoint          string
	PrivateKey        string
	ContractAddress   string
	ProposingInterval time.Duration
	EthClientTimeout  time.Duration
}

type ProposerMetrics interface {
	metrics.BasicMetrics
	RecordProposerTxSent(ctx context.Context, proposalData *scTypes.ProposalData)
}

func NewDefaultProposerParams() *ProposerParams {
	return &ProposerParams{
		Endpoint:          "http://rpc2.sepolia.org",
		PrivateKey:        "0000000000000000000000000000000000000000000000000000000000000001",
		ContractAddress:   "0x796baf7E572948CD0cbC374f345963bA433b47a2",
		ProposingInterval: 10 * time.Second,
		EthClientTimeout:  10 * time.Second,
	}
}

func NewProposer(
	ctx context.Context,
	params *ProposerParams,
	blockStorage storage.BlockStorage,
	ethClient rollupcontract.EthClient,
	metrics ProposerMetrics,
	logger zerolog.Logger,
) (*Proposer, error) {
	retryRunner := common.NewRetryRunner(
		common.RetryConfig{
			ShouldRetry: common.LimitRetries(5),
			NextDelay:   common.DelayExponential(100*time.Millisecond, time.Second),
		},
		logger,
	)

	rollupContract, err := rollupcontract.NewWrapper(ctx, params.ContractAddress, params.PrivateKey, ethClient, params.EthClientTimeout, logger)
	if err != nil {
		return nil, err
	}

	p := Proposer{
		blockStorage:   blockStorage,
		rollupContract: rollupContract,
		params:         params,
		retryRunner:    retryRunner,
		metrics:        metrics,
		logger:         logger,
	}

	return &p, nil
}

func (p *Proposer) Run(ctx context.Context) error {
	shouldResetStorage, err := p.initializeProvedStateRoot(ctx)
	if err != nil {
		return err
	}

	if shouldResetStorage {
		p.logger.Warn().Msg("resetting TaskStorage and BlockStorage")
		// todo: reset TaskStorage and BlockStorage before starting Aggregator, TaskScheduler and TaskListener
	}

	concurrent.RunTickerLoop(ctx, p.params.ProposingInterval,
		func(ctx context.Context) {
			if err := p.proposeNextBlock(ctx); err != nil {
				p.logger.Error().Err(err).Msg("error during proved blocks proposing")
				p.metrics.RecordError(ctx, "proposer")
				return
			}
		},
	)

	return nil
}

func (p *Proposer) initializeProvedStateRoot(ctx context.Context) (shouldResetStorage bool, err error) {
	storedStateRoot, err := p.blockStorage.TryGetProvedStateRoot(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if proved state root is initialized: %w", err)
	}

	latestStateRoot, err := p.getLatestProvedStateRoot(ctx)
	if err != nil {
		// TODO return error after enable local L1 endpoint
		p.logger.Error().Err(err).Msg("failed get current contract state root, set 0")
	}

	switch {
	case storedStateRoot == nil:
		p.logger.Info().
			Stringer("latestStateRoot", latestStateRoot).
			Msg("proved state root is not initialized, value from L1 will be used")
	case *storedStateRoot != latestStateRoot:
		p.logger.Warn().
			Stringer("storedStateRoot", storedStateRoot).
			Stringer("latestStateRoot", latestStateRoot).
			Msg("proved state root value is invalid, local storage will be reset")
		shouldResetStorage = true
	default:
		p.logger.Info().Stringer("stateRoot", storedStateRoot).Msg("proved state root value is valid")
	}

	if storedStateRoot == nil || *storedStateRoot != latestStateRoot {
		err = p.blockStorage.SetProvedStateRoot(ctx, latestStateRoot)
		if err != nil {
			return false, fmt.Errorf("failed set proved state root: %w", err)
		}
	}

	p.logger.Info().
		Stringer("stateRoot", latestStateRoot).
		Msg("proposer is initialized")
	return shouldResetStorage, nil
}

func (p *Proposer) proposeNextBlock(ctx context.Context) error {
	data, err := p.blockStorage.TryGetNextProposalData(ctx)
	if err != nil {
		return fmt.Errorf("failed get next block to propose: %w", err)
	}
	if data == nil {
		p.logger.Debug().Msg("no block to propose")
		return nil
	}

	err = p.sendProof(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to send proof to L1 for block with hash=%s: %w", data.MainShardBlockHash, err)
	}

	blockId := scTypes.NewBlockId(types.MainShardId, data.MainShardBlockHash)
	err = p.blockStorage.SetBlockAsProposed(ctx, blockId)
	if err != nil {
		return fmt.Errorf("failed set block with hash=%s as proposed: %w", data.MainShardBlockHash, err)
	}
	return nil
}

func (p *Proposer) getLatestProvedStateRoot(ctx context.Context) (common.Hash, error) {
	var finalizedBatchIndex string
	err := p.retryRunner.Do(ctx, func(context.Context) error {
		var err error
		finalizedBatchIndex, err = p.rollupContract.FinalizedBatchIndex(ctx)
		return err
	})
	if err != nil {
		return common.EmptyHash, err
	}

	var latestProvedState [32]byte
	err = p.retryRunner.Do(ctx, func(context.Context) error {
		var err error
		latestProvedState, err = p.rollupContract.StateRoots(ctx, finalizedBatchIndex)
		return err
	})

	return latestProvedState, err
}

func (p *Proposer) commitBatch(ctx context.Context, blobs []kzg4844.Blob, batchIndexInBlobStorage string) (*ethtypes.Transaction, bool, error) {
	var tx *ethtypes.Transaction
	batchTxSkipped := false
	err := p.retryRunner.Do(ctx, func(context.Context) error {
		var err error
		tx, err = p.rollupContract.CommitBatch(ctx, blobs, batchIndexInBlobStorage)
		if errors.Is(err, rollupcontract.ErrBatchAlreadyCommitted) {
			p.logger.Warn().Msg("batch is already committed, skipping blob tx")
			batchTxSkipped = true
			return nil
		}
		return err
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to upload blob: %w", err)
	}

	if !batchTxSkipped {
		p.logger.Info().
			Hex("txHash", tx.Hash().Bytes()).
			Int("gasLimit", int(tx.Gas())).
			Int("blobGasLimit", int(tx.BlobGas())).
			Int("cost", int(tx.Cost().Uint64())).
			Any("blobHases", tx.BlobHashes()).
			Msg("blob transaction sent")

		receipt, err := p.rollupContract.WaitForReceipt(ctx, tx.Hash())
		if err != nil {
			return nil, false, err
		}
		if receipt == nil {
			return nil, false, errors.New("CommitBatch tx mining timout exceeded")
		}
		if receipt.Status != ethtypes.ReceiptStatusSuccessful {
			return nil, false, errors.New("CommitBatch tx failed")
		}
	}

	return tx, batchTxSkipped, nil
}

func (p *Proposer) updateState(ctx context.Context, tx *ethtypes.Transaction, data *scTypes.ProposalData, batchIndexInBlobStorage string) error {
	blobTxSidecar := tx.BlobTxSidecar()
	dataProofs, err := rollupcontract.ComputeDataProofs(blobTxSidecar)
	if err != nil {
		return err
	}

	// TODO: populate with actual data
	validityProof := []byte{0x0A, 0x0B, 0x0C}

	p.logger.Info().
		Stringer("blockHash", data.MainShardBlockHash).
		Int("txCount", len(data.Transactions)).
		Msg("calling UpdateState L1 method")

	updateTxSkipped := false
	err = p.retryRunner.Do(ctx, func(context.Context) error {
		var err error
		tx, err = p.rollupContract.UpdateState(
			ctx,
			batchIndexInBlobStorage,
			data.OldProvedStateRoot,
			data.NewProvedStateRoot,
			dataProofs,
			blobTxSidecar.BlobHashes(),
			validityProof,
			rollupcontract.INilRollupPublicDataInfo{
				Placeholder1: []byte{0x07, 0x08, 0x09},
				Placeholder2: []byte{0x07, 0x08, 0x09},
			},
		)
		if errors.Is(err, rollupcontract.ErrBatchAlreadyFinalized) {
			p.logger.Warn().Msg("batch is already committed, skipping UpdateState tx")
			updateTxSkipped = true
			return nil
		}
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	if !updateTxSkipped {
		p.logger.Info().
			Hex("txHash", tx.Hash().Bytes()).
			Int("gasLimit", int(tx.Gas())).
			Int("cost", int(tx.Cost().Uint64())).
			Msg("UpdateState transaction sent")

		p.metrics.RecordProposerTxSent(ctx, data)
	}

	return nil
}

func (p *Proposer) sendProof(ctx context.Context, data *scTypes.ProposalData) error {
	// TODO: populate with actual data
	blobs := []kzg4844.Blob{{0x01}, {0x02}, {0x03}}
	batchIndexInBlobStorage := "0x0000000000000000000000000000000000000000000000000000000000000001"

	tx, _, err := p.commitBatch(ctx, blobs, batchIndexInBlobStorage)
	if err != nil {
		return err
	}

	if err := p.updateState(ctx, tx, data, batchIndexInBlobStorage); err != nil {
		return err
	}

	return nil
}
