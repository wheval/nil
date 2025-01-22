package tests

import (
	"testing"
	"time"

	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	"github.com/NilFoundation/nil/nil/tests"
	"github.com/stretchr/testify/suite"
)

type SuiteConsensus struct {
	tests.ShardedSuite
}

func (s *SuiteConsensus) SetupTest() {
	nShards := uint32(3)

	s.StartShardAllValidators(&nilservice.Config{
		NShards:              nShards,
		CollatorTickPeriodMs: 200,
	}, 10625)
}

func (s *SuiteConsensus) TearDownTest() {
	s.Cancel()
}

func (s *SuiteConsensus) TestConsensus() {
	// Check block id grows
	for shardId, shard := range s.Shards {
		block, err := shard.Client.GetBlock(s.Context, types.ShardId(shardId), "latest", true)
		s.Require().NoError(err)

		s.Require().Eventually(func() bool {
			newBlock, err := shard.Client.GetBlock(s.Context, types.ShardId(shardId), "latest", true)
			s.Require().NoError(err)
			return newBlock != nil && newBlock.Number > block.Number
		}, 30*time.Second, 1*time.Second)
	}

	// Check that all validators have the same block
	blocks := make([]*jsonrpc.RPCBlock, 0, len(s.Shards))
	for id := range s.Shards {
		block, err := s.Shards[0].Client.GetBlock(s.Context, types.ShardId(id), "latest", true)
		s.Require().NoError(err)
		blocks = append(blocks, block)
	}

	for _, block := range blocks {
		for id := range s.Shards {
			instanceBlock, err := s.Shards[id].Client.GetBlock(s.Context, block.ShardId, uint64(block.Number), false)
			s.Require().NoError(err)
			s.Equal(block.Hash, instanceBlock.Hash)
		}
	}
}

func TestConsensus(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(SuiteConsensus))
}
