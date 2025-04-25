package fetching

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/concurrent"
	"github.com/NilFoundation/nil/nil/common/logging"
	coreTypes "github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/synccommittee/core/batches/blob"
	"github.com/NilFoundation/nil/nil/services/synccommittee/core/batches/encode"
	v1 "github.com/NilFoundation/nil/nil/services/synccommittee/core/batches/encode/v1"
	"github.com/NilFoundation/nil/nil/services/synccommittee/core/reset"
	"github.com/NilFoundation/nil/nil/services/synccommittee/core/rollupcontract"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/metrics"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/srv"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/storage"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/jonboulle/clockwork"
)

type AggregatorMetrics interface {
	metrics.BasicMetrics
	RecordBatchCreated(ctx context.Context, batch *types.BlockBatch)
}

type AggregatorTaskStorage interface {
	AddTaskEntries(ctx context.Context, tasks ...*types.TaskEntry) error
}

type AggregatorBlockStorage interface {
	GetLatestFetched(ctx context.Context) (types.BlockRefs, error)
	TryGetProvedStateRoot(ctx context.Context) (*common.Hash, error)
	TryGetLatestBatch(ctx context.Context) (*types.BlockBatch, error)
	SetBlockBatch(ctx context.Context, batch *types.BlockBatch) error
	GetFreeSpaceBatchCount(ctx context.Context) (uint32, error)
	SetProvedStateRoot(ctx context.Context, stateRoot common.Hash) error
}

type AggregatorConfig struct {
	RpcPollingInterval time.Duration `yaml:"pollingDelay,omitempty"`
	MaxBlobsInTx       uint          `yaml:"-"`
}

func NewAggregatorConfig(rpcPollingInterval time.Duration) AggregatorConfig {
	return AggregatorConfig{
		RpcPollingInterval: rpcPollingInterval,
		MaxBlobsInTx:       6,
	}
}

func NewDefaultAggregatorConfig() AggregatorConfig {
	return NewAggregatorConfig(time.Second)
}

type aggregator struct {
	rpcClient       RpcBlockFetcher
	blockStorage    AggregatorBlockStorage
	taskStorage     AggregatorTaskStorage
	subgraphFetcher *subgraphFetcher
	batchEncoder    encode.BatchEncoder
	blobBuilder     blob.Builder
	rollupContract  rollupcontract.Wrapper
	resetter        *reset.StateResetLauncher
	clock           clockwork.Clock
	metrics         AggregatorMetrics
	config          AggregatorConfig
	workerAction    *concurrent.Suspendable
	logger          logging.Logger
}

func NewAggregator(
	rpcClient RpcBlockFetcher,
	blockStorage AggregatorBlockStorage,
	taskStorage AggregatorTaskStorage,
	resetter *reset.StateResetLauncher,
	rollupContractWrapper rollupcontract.Wrapper,
	clock clockwork.Clock,
	logger logging.Logger,
	metrics AggregatorMetrics,
	config AggregatorConfig,
) *aggregator {
	agg := &aggregator{
		rpcClient:       rpcClient,
		blockStorage:    blockStorage,
		taskStorage:     taskStorage,
		subgraphFetcher: newSubgraphFetcher(rpcClient, logger),
		batchEncoder:    v1.NewEncoder(logger),
		blobBuilder:     blob.NewBuilder(),
		rollupContract:  rollupContractWrapper,
		resetter:        resetter,
		clock:           clock,
		metrics:         metrics,
		config:          config,
	}

	agg.workerAction = concurrent.NewSuspendable(agg.runIteration, config.RpcPollingInterval)
	agg.logger = srv.WorkerLogger(logger, agg)
	return agg
}

func (agg *aggregator) Name() string {
	return "aggregator"
}

func (agg *aggregator) Run(ctx context.Context, started chan<- struct{}) error {
	agg.logger.Info().Msg("Starting block fetching")

	latestStateRoot, err := agg.blockStorage.TryGetProvedStateRoot(ctx)
	if err != nil {
		return err
	}
	if latestStateRoot == nil {
		// root not initialized in current database
		agg.logger.Warn().
			Msg("L1 state root is not initialized, trying to use storage state root")
		if err := agg.syncLatestFinalizedRoot(ctx); err != nil {
			return fmt.Errorf("failed to set proved state root: %w", err)
		}
	}

	err = agg.workerAction.Run(ctx, started)

	if err == nil || errors.Is(err, context.Canceled) {
		agg.logger.Info().Msg("Block fetching stopped")
	} else {
		agg.logger.Error().Err(err).Msg("Error running aggregator, stopped")
	}

	return err
}

func (agg *aggregator) Pause(ctx context.Context) error {
	paused, err := agg.workerAction.Pause(ctx)
	if err != nil {
		return err
	}
	if paused {
		agg.logger.Info().Msg("Block fetching paused")
	} else {
		agg.logger.Warn().Msg("Block fetching already paused")
	}
	return nil
}

func (agg *aggregator) Resume(ctx context.Context) error {
	resumed, err := agg.workerAction.Resume(ctx)
	if err != nil {
		return err
	}
	if resumed {
		agg.logger.Info().Msg("Block fetching resumed")
	} else {
		agg.logger.Warn().Msg("Block fetching already running")
	}
	return nil
}

func (agg *aggregator) runIteration(ctx context.Context) {
	err := agg.processBlocksAndHandleErr(ctx)
	if err != nil {
		agg.metrics.RecordError(ctx, agg.Name())
	}
}

// processBlocksAndHandleErr fetches and processes new blocks for all shards.
// It handles the overall flow of block synchronization and proof creation.
func (agg *aggregator) processBlocksAndHandleErr(ctx context.Context) error {
	err := agg.processBlockRange(ctx)
	return agg.handleProcessingErr(ctx, err)
}

func (agg *aggregator) handleProcessingErr(ctx context.Context, err error) error {
	switch {
	case err == nil:
		return nil

	case errors.Is(err, types.ErrBlockMismatch):
		agg.logger.Warn().Err(err).Msg("Block mismatch detected, resetting state")
		if err := agg.resetter.LaunchResetToL1WithSuspension(ctx, agg); err != nil {
			return fmt.Errorf("error resetting state: %w", err)
		}
		return nil

	case errors.Is(err, storage.ErrStateRootNotInitialized):
		agg.logger.Warn().Err(err).Msg("State root not initialized, trying to fetch it from L1")
		if err := agg.syncLatestFinalizedRoot(ctx); err != nil {
			return fmt.Errorf("failed to set proved state root: %w", err)
		}
		return nil

	case errors.Is(err, storage.ErrCapacityLimitReached):
		agg.logger.Info().Err(err).Msg("Storage capacity limit reached, skipping")
		return nil

	case errors.Is(err, context.Canceled):
		agg.logger.Info().Err(err).Msg("Block processing cancelled")
		return err

	default:
		agg.logger.Error().Err(err).Msg("Error processing blocks")
		return err
	}
}

// processBlockRange handles the processing of new blocks for the main shard.
// It fetches new blocks, updates the storage, and records relevant metrics.
func (agg *aggregator) processBlockRange(ctx context.Context) error {
	startingBlockRef, err := agg.getStartingBlockRef(ctx)
	if err != nil {
		return err
	}

	latestBlockRef, err := agg.getLatestBlockRef(ctx)
	if err != nil {
		return err
	}

	maxNumBatches, err := agg.blockStorage.GetFreeSpaceBatchCount(ctx)
	if err != nil {
		return err
	}
	if maxNumBatches == 0 {
		return fmt.Errorf("%w, cannot fetch blocks", storage.ErrCapacityLimitReached)
	}

	fetchingRange, err := types.GetBlocksFetchingRange(*startingBlockRef, *latestBlockRef, maxNumBatches)
	if err != nil {
		return err
	}

	if fetchingRange == nil {
		agg.logger.Debug().
			Stringer(logging.FieldShardId, coreTypes.MainShardId).
			Stringer(logging.FieldBlockNumber, latestBlockRef.Number).
			Msg("No new blocks to fetch")
		return nil
	}

	return agg.fetchAndProcessBlocks(ctx, *fetchingRange)
}

// fetchLatestBlocks retrieves the latest block for main shard
func (agg *aggregator) getLatestBlockRef(ctx context.Context) (*types.BlockRef, error) {
	block, err := agg.rpcClient.GetBlock(ctx, coreTypes.MainShardId, "latest", false)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest block from shard %d: %w", coreTypes.MainShardId, err)
	}
	if block == nil {
		return nil, fmt.Errorf("%w: latest main block not found in chain", types.ErrBlockNotFound)
	}
	blockRef := types.BlockToRef(block)
	return &blockRef, nil
}

// getStartingBlockRef retrieves the starting point for the next fetching iteration,
// prioritizing the latest fetched main shard block if available.
// If `latestFetched` value is not defined, method uses `latestProvedStateRoot`.
// If neither of the two values is defined, method returns an error.
func (agg *aggregator) getStartingBlockRef(ctx context.Context) (*types.BlockRef, error) {
	latestFetched, err := agg.blockStorage.GetLatestFetched(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading latest fetched block for the main shard: %w", err)
	}
	if mainRef := latestFetched.TryGetMain(); mainRef != nil {
		// checking if `latestFetched` still exists on L2 side
		if _, err := agg.getBlockRef(ctx, mainRef.ShardId, mainRef.Hash); err != nil {
			return nil, fmt.Errorf("fetched block check error: %w", err)
		}

		return mainRef, nil
	}

	agg.logger.Debug().Msg("No blocks fetched yet, latest proved state root value will be used")

	latestProvedRoot, err := agg.blockStorage.TryGetProvedStateRoot(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading latest proved state root: %w", err)
	}
	if latestProvedRoot == nil {
		return nil, storage.ErrStateRootNotInitialized
	}

	ref, err := agg.getBlockRef(ctx, coreTypes.MainShardId, *latestProvedRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to get proved block ref: %w", err)
	}
	return ref, nil
}

// getBlockRef retrieves the block reference to the specified block using the RPC client.
// If block does not exist, method returns types.ErrBlockMismatch.
func (agg *aggregator) getBlockRef(
	ctx context.Context,
	shard coreTypes.ShardId,
	hash common.Hash,
) (*types.BlockRef, error) {
	rpcBlock, err := agg.rpcClient.GetBlock(ctx, shard, hash, false)
	if err != nil {
		return nil, fmt.Errorf("%w: error fetching block, shard=%d, hash=%s", err, shard, hash)
	}
	if rpcBlock == nil {
		return nil, fmt.Errorf("%w: block not found in chain, shard=%d, hash=%s", types.ErrBlockMismatch, shard, hash)
	}

	ref := types.BlockToRef(rpcBlock)
	return &ref, nil
}

// fetchAndProcessBlocks retrieves a range of blocks for a main shard, stores them, creates proof tasks
func (agg *aggregator) fetchAndProcessBlocks(ctx context.Context, blocksRange types.BlocksRange) error {
	shardId := coreTypes.MainShardId
	const requestBatchSize = 20
	mainShardBlocks, err := agg.rpcClient.GetBlocksRange(
		ctx, shardId, blocksRange.Start, blocksRange.End+1, true, requestBatchSize,
	)
	if err != nil {
		return fmt.Errorf(
			"error fetching blocks from shard %d in range [%d, %d]: %w",
			shardId, blocksRange.Start, blocksRange.End, err,
		)
	}

	for _, mainShardBlock := range mainShardBlocks {
		blockBatch, err := agg.createBlockBatch(ctx, mainShardBlock)
		if err != nil {
			return fmt.Errorf("error creating batch, mainHash=%s: %w", mainShardBlock.Hash, err)
		}

		if err := agg.handleBlockBatch(ctx, blockBatch); err != nil {
			return fmt.Errorf("error handing batch, mainHash=%s: %w", mainShardBlock.Hash, err)
		}
	}

	return nil
}

func (agg *aggregator) createBlockBatch(
	ctx context.Context,
	mainShardBlock *types.Block,
) (*types.BlockBatch, error) {
	latestBatch, err := agg.blockStorage.TryGetLatestBatch(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading latest batch: %w", err)
	}
	var latestBatchId *types.BatchId
	if latestBatch != nil {
		latestBatchId = &latestBatch.Id
	}

	latestFetched, err := agg.blockStorage.GetLatestFetched(ctx)
	if err != nil {
		return nil, err
	}

	subgraph, err := agg.subgraphFetcher.FetchSubgraph(ctx, mainShardBlock, latestFetched)
	if err != nil {
		return nil, err
	}

	return types.NewBlockBatch(latestBatchId).WithAddedBlocks(subgraph)
}

// handleBlockBatch checks the validity of a block and stores it if valid.
func (agg *aggregator) handleBlockBatch(ctx context.Context, batch *types.BlockBatch) error {
	latestFetched, err := agg.blockStorage.GetLatestFetched(ctx)
	if err != nil {
		return fmt.Errorf("error reading latest fetched block from storage: %w", err)
	}

	mainRef := latestFetched.TryGetMain()
	if err := mainRef.ValidateNext(batch.EarliestMainBlock()); err != nil {
		return err
	}

	sidecar, dataProofs, err := agg.prepareForBatchCommit(ctx, batch)
	if err != nil {
		return err
	}
	batch = batch.WithDataProofs(dataProofs)

	if err := agg.blockStorage.SetBlockBatch(ctx, batch); err != nil {
		return fmt.Errorf("error storing block batch, latestMainHash=%s: %w", batch.LatestMainBlock().Hash, err)
	}

	if err := agg.rollupContract.CommitBatch(ctx, sidecar, batch.Id.String()); err != nil {
		return agg.handleCommitBatchError(ctx, batch, err)
	}

	if err := agg.createProofTasks(ctx, batch); err != nil {
		return fmt.Errorf("error creating proof tasks, latestMainHash=%s: %w", batch.LatestMainBlock().Hash, err)
	}

	agg.metrics.RecordBatchCreated(ctx, batch)
	return nil
}

func (agg *aggregator) handleCommitBatchError(ctx context.Context, batch *types.BlockBatch, err error) error {
	switch {
	case errors.Is(err, rollupcontract.ErrBatchAlreadyCommitted) ||
		errors.Is(err, rollupcontract.ErrBatchAlreadyFinalized):
		// for some reason, we attempted to prove a batch that has already been proved,
		// sync the latest proved root with the L1 contract.
		agg.logger.Warn().Stringer(logging.FieldBatchId, batch.Id).
			Err(err).Msg("batch is already committed, resetting state with L1")
		if err := agg.resetter.LaunchResetToL1WithSuspension(ctx, agg); err != nil {
			return fmt.Errorf("error resetting state from L1, latestMainHash=%s: %w",
				batch.LatestMainBlock().Hash, err)
		}
	case errors.Is(err, rollupcontract.ErrInvalidBatchIndex) ||
		errors.Is(err, rollupcontract.ErrInvalidVersionedHash):
		// data was corrupted or initially created in a wrong way
		// NOTE: this shouldn't happen in prod setting
		agg.logger.Warn().Stringer(logging.FieldBatchId, batch.Id).
			Err(err).Msg("data was corrupted or initially created in a wrong way")
		if err := agg.resetter.LaunchResetToL1WithSuspension(ctx, agg); err != nil {
			return fmt.Errorf("error resetting state from L1, latestMainHash=%s: %w",
				batch.LatestMainBlock().Hash, err)
		}
	default:
		return err
	}
	return nil
}

// createProofTask generates proof task for block batch
func (agg *aggregator) createProofTasks(ctx context.Context, batch *types.BlockBatch) error {
	currentTime := agg.clock.Now()
	proofTask, err := batch.CreateProofTask(currentTime)
	if err != nil {
		return fmt.Errorf("error creating proof tasks, latestMainHash=%s: %w", batch.LatestMainBlock().Hash, err)
	}

	if err := agg.taskStorage.AddTaskEntries(ctx, proofTask); err != nil {
		return fmt.Errorf("error adding task entries, latestMainHash=%s: %w", batch.LatestMainBlock().Hash, err)
	}

	agg.logger.Debug().
		Stringer(logging.FieldBatchId, batch.Id).
		Msgf("Created proof task, latestMainHash=%s", batch.LatestMainBlock().Hash)

	return nil
}

func (agg *aggregator) prepareForBatchCommit(
	ctx context.Context, batch *types.BlockBatch,
) (*ethtypes.BlobTxSidecar, types.DataProofs, error) {
	var binTransactions bytes.Buffer
	if err := agg.batchEncoder.Encode(types.NewPrunedBatch(batch), &binTransactions); err != nil {
		return nil, nil, err
	}
	agg.logger.Debug().Int("compressed_batch_len", binTransactions.Len()).Msg("encoded transaction")

	blobs, err := agg.blobBuilder.MakeBlobs(&binTransactions, agg.config.MaxBlobsInTx)
	if err != nil {
		return nil, nil, err
	}

	return agg.rollupContract.PrepareBlobs(ctx, blobs)
}

// getLatestFinalizedRootFromL1 attempts to retrieve the finalized root from the following sources,
// in order of priority:
// 1. L1 contract
// 2. Genesis block (fallback)
func (agg *aggregator) getLatestFinalizedRootFromL1(ctx context.Context) (common.Hash, error) {
	agg.logger.Info().Msg("syncing state with L1")

	latestStateRoot, err := agg.rollupContract.LatestFinalizedStateRoot(ctx)
	if err != nil {
		return common.EmptyHash, err
	}
	if latestStateRoot != common.EmptyHash {
		return latestStateRoot, nil
	}

	agg.logger.Warn().
		Msg("storage state root is not initialized, genesis state root will be used")
	genesisBlock, err := agg.rpcClient.GetBlock(ctx, coreTypes.MainShardId, "earliest", false)
	if err != nil {
		return common.EmptyHash, err
	}
	latestStateRoot = genesisBlock.Hash

	return latestStateRoot, nil
}

func (agg *aggregator) syncLatestFinalizedRoot(ctx context.Context) error {
	latestStateRoot, err := agg.getLatestFinalizedRootFromL1(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest finalized state root: %w", err)
	}
	if err := agg.blockStorage.SetProvedStateRoot(ctx, latestStateRoot); err != nil {
		return fmt.Errorf("failed to set proved state root: %w", err)
	}
	agg.logger.Warn().Msg("Proved state root set")
	return nil
}
