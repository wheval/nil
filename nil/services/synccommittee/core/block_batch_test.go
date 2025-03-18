package core

import (
	"testing"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/testaide"
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"github.com/stretchr/testify/suite"
)

type BlockBatchTestSuite struct {
	suite.Suite
}

func TestBlockBatchTestSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(BlockBatchTestSuite))
}

func (s *BlockBatchTestSuite) TestNewBlockBatch() {
	validBatch := testaide.NewBlockBatch(testaide.ShardsCount)

	notReadyBatch := testaide.NewBlockBatch(testaide.ShardsCount)
	notReadyBatch.MainShardBlock.ChildBlocks[0] = common.EmptyHash
	notReadyBatch.ChildBlocks[0] = nil

	nilChildBatch := testaide.NewBlockBatch(testaide.ShardsCount)
	nilChildBatch.ChildBlocks[1] = nil

	redundantChildBatch := testaide.NewBlockBatch(testaide.ShardsCount)
	redundantChildBatch.ChildBlocks = append(redundantChildBatch.ChildBlocks, testaide.NewExecutionShardBlock())

	hashMismatchBatch := testaide.NewBlockBatch(testaide.ShardsCount)
	hashMismatchBatch.ChildBlocks[2].Hash = testaide.RandomHash()

	testCases := []struct {
		name           string
		mainShardBlock *jsonrpc.RPCBlock
		childBlocks    []*jsonrpc.RPCBlock
		errPredicate   func(error)
	}{
		{
			name:           "Valid_Batch_No_Error",
			mainShardBlock: validBatch.MainShardBlock,
			childBlocks:    validBatch.ChildBlocks,
			errPredicate:   func(err error) { s.Require().NoError(err) },
		},
		{
			name:           "Nil_Main_Shard_Block",
			mainShardBlock: nil,
			childBlocks:    []*jsonrpc.RPCBlock{},
			errPredicate:   func(err error) { s.Require().ErrorContains(err, "mainShardBlock") },
		},
		{
			name:           "Valid_Main_Shard_Block_Nil_Child_Blocks",
			mainShardBlock: testaide.NewMainShardBlock(),
			childBlocks:    nil,
			errPredicate:   func(err error) { s.Require().ErrorContains(err, "childBlocks") },
		},
		{
			name:           "Not_Ready_Batch_Child_Hash_Is_Empty",
			mainShardBlock: notReadyBatch.MainShardBlock,
			childBlocks:    notReadyBatch.ChildBlocks,
			errPredicate:   func(err error) { s.Require().ErrorIs(err, types.ErrBatchNotReady) },
		},
		{
			name:           "Valid_Main_Shard_Block_Nil_Child_Block",
			mainShardBlock: nilChildBatch.MainShardBlock,
			childBlocks:    nilChildBatch.ChildBlocks,
			errPredicate: func(err error) {
				s.Require().NotErrorIs(err, types.ErrBatchNotReady)
				s.Require().ErrorContains(err, "childBlocks[1] cannot be nil")
			},
		},
		{
			name:           "Block_Is_Not_From_The_Main_Shard",
			mainShardBlock: testaide.NewExecutionShardBlock(),
			childBlocks:    []*jsonrpc.RPCBlock{},
			errPredicate: func(err error) {
				s.Require().ErrorContains(err, "mainShardBlock is not from the main shard")
			},
		},
		{
			name:           "Redundant_Child_Block",
			mainShardBlock: redundantChildBatch.MainShardBlock,
			childBlocks:    redundantChildBatch.ChildBlocks,
			errPredicate:   func(err error) { s.Require().ErrorContains(err, "have different length") },
		},
		{
			name:           "Child_Hash_Mismatch",
			mainShardBlock: hashMismatchBatch.MainShardBlock,
			childBlocks:    hashMismatchBatch.ChildBlocks,
			errPredicate: func(err error) {
				s.Require().ErrorContains(err, "childBlocks[2].Hash != mainShardBlock.ChildBlocks[2]")
			},
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			parentBatchId := types.NewBatchId()
			batch, err := types.NewBlockBatch(&parentBatchId, testCase.mainShardBlock, testCase.childBlocks)
			testCase.errPredicate(err)

			if err != nil {
				s.Require().Nil(batch)
				return
			}

			s.Require().NotNil(batch)
			s.Require().Equal(testCase.mainShardBlock, batch.MainShardBlock)
			s.Require().Equal(testCase.childBlocks, batch.ChildBlocks)
		})
	}
}

func (s *BlockBatchTestSuite) TestCreateProofTask() {
	const childBLockCount = 4
	batch := testaide.NewBlockBatch(childBLockCount)

	taskEntry, err := batch.CreateProofTask(testaide.Now)
	s.Require().NoError(err)

	task := taskEntry.Task
	s.Require().Equal(types.ProofBatch, task.TaskType)
	s.Require().Equal(batch.Id, task.BatchId)
	s.Require().Equal(batch.MainShardBlock.Hash, task.BlockIds[0].Hash)
	s.Require().Nil(task.ParentTaskId)

	for i, childBlock := range batch.ChildBlocks {
		// `testaide.NewBlockBatch(n)` creates one main shard block and `n` child blocks, each for different shard
		taskChildBlock := task.BlockIds[i+1]

		s.Require().Equal(childBlock.Hash, taskChildBlock.Hash)
	}
}
