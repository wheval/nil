package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	scTypes "github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"github.com/jonboulle/clockwork"
)

type BlockStorageMetrics interface {
	RecordBatchProved(ctx context.Context)
}

type BlockStorageConfig struct {
	// StoredBatchesLimit defines the maximum number of stored batches.
	// If the capacity limit is reached, method BlockStorage.SetBlockBatch returns ErrCapacityLimitReached error.
	StoredBatchesLimit uint32
}

func NewBlockStorageConfig(storedBatchesLimit uint32) BlockStorageConfig {
	return BlockStorageConfig{
		StoredBatchesLimit: storedBatchesLimit,
	}
}

func DefaultBlockStorageConfig() BlockStorageConfig {
	return NewBlockStorageConfig(100)
}

type BlockStorage struct {
	commonStorage
	config  BlockStorageConfig
	clock   clockwork.Clock
	metrics BlockStorageMetrics

	ops struct {
		batchOp
		batchCountOp
		batchLatestOp
		blockOp
		blockLatestFetchedOp
		stateRootOp
	}
}

func NewBlockStorage(
	database db.DB,
	config BlockStorageConfig,
	clock clockwork.Clock,
	metrics BlockStorageMetrics,
	logger logging.Logger,
) *BlockStorage {
	return &BlockStorage{
		commonStorage: makeCommonStorage(
			database,
			logger,
			common.DoNotRetryIf(
				scTypes.ErrBatchMismatch, scTypes.ErrBlockNotFound, scTypes.ErrBatchNotFound, scTypes.ErrBatchNotProved,
				ErrStateRootNotInitialized,
			),
		),
		config:  config,
		clock:   clock,
		metrics: metrics,
	}
}

func (bs *BlockStorage) TryGetProvedStateRoot(ctx context.Context) (*common.Hash, error) {
	tx, err := bs.database.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	return bs.ops.getProvedStateRoot(tx)
}

func (bs *BlockStorage) SetProvedStateRoot(ctx context.Context, stateRoot common.Hash) error {
	if stateRoot == common.EmptyHash {
		return errors.New("state root cannot be empty")
	}

	tx, err := bs.database.CreateRwTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := bs.ops.putProvedStateRoot(tx, stateRoot); err != nil {
		return err
	}

	return bs.commit(tx)
}

// TryGetLatestBatchId retrieves the ID of the latest created batch
// or returns nil if:
// a) No batches have been created yet, or
// b) A full storage reset (starting from the first batch) has been triggered.
func (bs *BlockStorage) TryGetLatestBatchId(ctx context.Context) (*scTypes.BatchId, error) {
	tx, err := bs.database.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	return bs.ops.getLatestBatchId(tx)
}

func (bs *BlockStorage) TryGetLatestFetched(ctx context.Context) (*scTypes.MainBlockRef, error) {
	tx, err := bs.database.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	lastFetched, err := bs.ops.getLatestFetchedMain(tx)
	if err != nil {
		return nil, err
	}

	return lastFetched, nil
}

func (bs *BlockStorage) TryGetBlock(ctx context.Context, id scTypes.BlockId) (*jsonrpc.RPCBlock, error) {
	tx, err := bs.database.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	entry, err := bs.ops.getBlock(tx, id, false)
	if err != nil || entry == nil {
		return nil, err
	}
	return &entry.Block, nil
}

func (bs *BlockStorage) GetFreeSpaceBatchCount(ctx context.Context) (uint32, error) {
	tx, err := bs.database.CreateRoTx(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	batchCount, err := bs.ops.getBatchesCount(tx)
	if err != nil {
		return 0, err
	}
	return bs.config.StoredBatchesLimit - batchCount, nil
}

func (bs *BlockStorage) SetBlockBatch(ctx context.Context, batch *scTypes.BlockBatch) error {
	if batch == nil {
		return errors.New("batch cannot be nil")
	}

	return bs.retryRunner.Do(ctx, func(ctx context.Context) error {
		return bs.setBlockBatchImpl(ctx, batch)
	})
}

func (bs *BlockStorage) setBlockBatchImpl(ctx context.Context, batch *scTypes.BlockBatch) error {
	tx, err := bs.database.CreateRwTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := bs.putBatchWithBlocks(tx, batch); err != nil {
		return err
	}

	if err := bs.setProposeParentHash(tx, batch.MainShardBlock); err != nil {
		return err
	}

	if err := bs.ops.updateLatestFetched(tx, batch.MainShardBlock); err != nil {
		return err
	}

	if err := bs.ops.updateLatestBatchId(tx, batch); err != nil {
		return err
	}

	return bs.commit(tx)
}

func (bs *BlockStorage) setProposeParentHash(tx db.RwTx, block *jsonrpc.RPCBlock) error {
	if block.ShardId != types.MainShardId {
		return nil
	}
	parentHash, err := bs.ops.getParentOfNextToPropose(tx)
	if err != nil {
		return err
	}
	if parentHash != nil {
		return nil
	}

	if block.Number > 0 && block.ParentHash.Empty() {
		return fmt.Errorf("block with hash=%s has empty parent hash", block.Hash.String())
	}

	bs.logger.Info().
		Stringer(logging.FieldBlockHash, block.Hash).
		Stringer("parentHash", block.ParentHash).
		Msg("block parent hash is not set, updating it")

	return bs.ops.setParentOfNextToPropose(tx, block.ParentHash)
}

func (bs *BlockStorage) SetBatchAsProved(ctx context.Context, batchId scTypes.BatchId) error {
	wasSet, err := bs.setBatchAsProvedImpl(ctx, batchId)
	if err != nil {
		return err
	}
	if wasSet {
		bs.metrics.RecordBatchProved(ctx)
	}
	return nil
}

func (bs *BlockStorage) setBatchAsProvedImpl(ctx context.Context, batchId scTypes.BatchId) (wasSet bool, err error) {
	tx, err := bs.database.CreateRwTx(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	entry, err := bs.ops.getBatch(tx, batchId)
	if err != nil {
		return false, err
	}

	if entry.IsProved {
		bs.logger.Debug().Stringer(logging.FieldBatchId, batchId).Msg("batch is already marked as proved")
		return false, nil
	}

	entry.IsProved = true
	if err := bs.ops.putBatch(tx, entry); err != nil {
		return false, err
	}

	if err := bs.commit(tx); err != nil {
		return false, err
	}

	return true, nil
}

func (bs *BlockStorage) TryGetNextProposalData(ctx context.Context) (*scTypes.ProposalData, error) {
	tx, err := bs.database.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	currentProvedStateRoot, err := bs.ops.getProvedStateRoot(tx)
	if err != nil {
		return nil, err
	}
	if currentProvedStateRoot == nil {
		return nil, ErrStateRootNotInitialized
	}

	parentHash, err := bs.ops.getParentOfNextToPropose(tx)
	if err != nil {
		return nil, err
	}

	if parentHash == nil {
		bs.logger.Debug().Msg("block parent hash is not set")
		return nil, nil
	}

	var proposalCandidate *batchEntry
	for entry, err := range bs.ops.getStoredBatchesSeq(tx) {
		if err != nil {
			return nil, err
		}
		if isValidProposalCandidate(entry, *parentHash) {
			proposalCandidate = entry
			break
		}
	}

	if proposalCandidate == nil {
		bs.logger.Debug().Stringer("parentHash", parentHash).Msg("no proved batch found")
		return nil, nil
	}

	return bs.createProposalDataTx(tx, proposalCandidate, currentProvedStateRoot)
}

func (bs *BlockStorage) createProposalDataTx(
	tx db.RoTx,
	proposalCandidate *batchEntry,
	currentProvedStateRoot *common.Hash,
) (*scTypes.ProposalData, error) {
	mainBlockEntry, err := bs.ops.getBlock(tx, proposalCandidate.MainBlockId, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get main block with id=%s: %w", proposalCandidate.MainBlockId, err)
	}

	transactions := scTypes.BlockTransactions(&mainBlockEntry.Block)

	for _, childId := range proposalCandidate.ExecBlockIds {
		childEntry, err := bs.ops.getBlock(tx, childId, true)
		if err != nil {
			return nil, fmt.Errorf("failed to get child block with id=%s: %w", childId, err)
		}

		blockTransactions := scTypes.BlockTransactions(&childEntry.Block)
		transactions = append(transactions, blockTransactions...)
	}

	return &scTypes.ProposalData{
		BatchId:            proposalCandidate.Id,
		MainShardBlockHash: mainBlockEntry.Block.Hash,
		Transactions:       transactions,
		OldProvedStateRoot: *currentProvedStateRoot,
		NewProvedStateRoot: mainBlockEntry.Block.Hash,
		MainBlockFetchedAt: mainBlockEntry.FetchedAt,
	}, nil
}

func (bs *BlockStorage) SetBatchAsProposed(ctx context.Context, id scTypes.BatchId) error {
	return bs.retryRunner.Do(ctx, func(ctx context.Context) error {
		return bs.setBatchAsProposedImpl(ctx, id)
	})
}

func (bs *BlockStorage) setBatchAsProposedImpl(ctx context.Context, id scTypes.BatchId) error {
	tx, err := bs.database.CreateRwTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	batch, err := bs.ops.getBatch(tx, id)
	if err != nil {
		return err
	}

	if !batch.IsProved {
		return fmt.Errorf("%w, id=%s", scTypes.ErrBatchNotProved, id)
	}

	mainShardEntry, err := bs.ops.getBlock(tx, batch.MainBlockId, true)
	if err != nil {
		return err
	}

	if err := bs.validateMainShardEntry(tx, mainShardEntry); err != nil {
		return err
	}

	if err := bs.deleteBatchWithBlocks(tx, batch); err != nil {
		return err
	}

	if err := bs.ops.putProvedStateRoot(tx, mainShardEntry.Block.Hash); err != nil {
		return err
	}
	if err := bs.ops.setParentOfNextToPropose(tx, mainShardEntry.Block.Hash); err != nil {
		return err
	}

	return bs.commit(tx)
}

func isValidProposalCandidate(batch *batchEntry, parentHash common.Hash) bool {
	return batch.IsProved && batch.MainParentBlockHash == parentHash
}

func (bs *BlockStorage) validateMainShardEntry(tx db.RoTx, entry *blockEntry) error {
	id := scTypes.IdFromBlock(&entry.Block)

	if entry.Block.ShardId != types.MainShardId {
		return fmt.Errorf("block with id=%s is not from main shard", id.String())
	}

	parentHash, err := bs.ops.getParentOfNextToPropose(tx)
	if err != nil {
		return err
	}
	if parentHash == nil {
		return errors.New("next to propose parent hash is not set")
	}

	if *parentHash != entry.Block.ParentHash {
		return fmt.Errorf(
			"parent's block hash=%s is not equal to the stored value=%s",
			entry.Block.ParentHash.String(),
			parentHash.String(),
		)
	}
	return nil
}

// ResetBatchesRange resets the block storage state starting from the batch with given ID:
//
//  1. Picks first main shard block [B] from the batch with the given ID.
//
//  2. Sets the latest fetched block reference to the parent of the block [B].
//     If the specified block is the first block in the chain, the new latest fetched value will be nil.
//
//  3. Deletes all main and corresponding exec shard blocks starting from the block [B].
func (bs *BlockStorage) ResetBatchesRange(
	ctx context.Context,
	firstBatchToPurge scTypes.BatchId,
) (purgedBatches []scTypes.BatchId, err error) {
	err = bs.retryRunner.Do(ctx, func(ctx context.Context) error {
		var err error
		purgedBatches, err = bs.resetBatchesPartialImpl(ctx, firstBatchToPurge)
		return err
	})
	return
}

func (bs *BlockStorage) resetBatchesPartialImpl(
	ctx context.Context,
	firstBatchToPurge scTypes.BatchId,
) (purgedBatches []scTypes.BatchId, err error) {
	tx, err := bs.database.CreateRwTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	startingBatch, err := bs.ops.getBatch(tx, firstBatchToPurge)
	if err != nil {
		return nil, err
	}

	if err := bs.resetToParent(tx, startingBatch); err != nil {
		return nil, err
	}

	for batch, err := range bs.ops.getBatchesSequence(tx, firstBatchToPurge) {
		if err != nil {
			return nil, err
		}

		if err := bs.deleteBatchWithBlocks(tx, batch); err != nil {
			return nil, err
		}

		purgedBatches = append(purgedBatches, batch.Id)
	}

	if err := bs.commit(tx); err != nil {
		return nil, err
	}

	return purgedBatches, nil
}

func (bs *BlockStorage) resetToParent(tx db.RwTx, batch *batchEntry) error {
	mainBlockEntry, err := bs.ops.getBlock(tx, batch.MainBlockId, true)
	if err != nil {
		return err
	}

	refToParent, err := scTypes.GetMainParentRef(&mainBlockEntry.Block)
	if err != nil {
		return fmt.Errorf("failed to get main block parent ref: %w", err)
	}
	if err := bs.ops.putLatestFetchedBlock(tx, types.MainShardId, refToParent); err != nil {
		return fmt.Errorf("failed to reset latest fetched block: %w", err)
	}
	if err := bs.ops.putLatestBatchId(tx, batch.ParentId); err != nil {
		return fmt.Errorf("failed to reset latest batch id: %w", err)
	}

	return nil
}

// ResetBatchesNotProved resets the block storage state:
//
//  1. Sets the latest fetched block reference to nil.
//
//  2. Deletes all main not yet proved blocks from the storage.
func (bs *BlockStorage) ResetBatchesNotProved(ctx context.Context) error {
	return bs.retryRunner.Do(ctx, func(ctx context.Context) error {
		return bs.resetBatchesNotProvedImpl(ctx)
	})
}

func (bs *BlockStorage) resetBatchesNotProvedImpl(ctx context.Context) error {
	tx, err := bs.database.CreateRwTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := bs.ops.putLatestFetchedBlock(tx, types.MainShardId, nil); err != nil {
		return fmt.Errorf("failed to reset latest fetched block: %w", err)
	}

	for batch, err := range bs.ops.getStoredBatchesSeq(tx) {
		if err != nil {
			return err
		}
		if batch.IsProved {
			continue
		}

		if err := bs.deleteBatchWithBlocks(tx, batch); err != nil {
			return err
		}
	}

	return bs.commit(tx)
}

func (bs *BlockStorage) putBatchWithBlocks(tx db.RwTx, batch *scTypes.BlockBatch) error {
	if err := bs.ops.addStoredCount(tx, 1, bs.config); err != nil {
		return err
	}

	if err := bs.ops.putBatchParentIndexEntry(tx, batch); err != nil {
		return err
	}

	currentTime := bs.clock.Now()

	entry := newBatchEntry(batch, currentTime)
	if err := bs.ops.putBatch(tx, entry); err != nil {
		return err
	}

	mainEntry := newBlockEntry(batch.MainShardBlock, batch, currentTime)
	if err := bs.ops.putBlockTx(tx, mainEntry); err != nil {
		return err
	}

	for _, childBlock := range batch.ChildBlocks {
		childEntry := newBlockEntry(childBlock, batch, currentTime)
		if err := bs.ops.putBlockTx(tx, childEntry); err != nil {
			return err
		}
	}

	return nil
}

func (bs *BlockStorage) deleteBatchWithBlocks(tx db.RwTx, batch *batchEntry) error {
	if err := bs.ops.addStoredCount(tx, -1, bs.config); err != nil {
		return err
	}

	if err := bs.ops.deleteBatch(tx, batch, bs.logger); err != nil {
		return err
	}

	if err := bs.ops.deleteBlock(tx, batch.MainBlockId, bs.logger); err != nil {
		return err
	}

	for _, childId := range batch.ExecBlockIds {
		if err := bs.ops.deleteBlock(tx, childId, bs.logger); err != nil {
			return err
		}
	}

	return nil
}
