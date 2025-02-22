package tests

import (
	"testing"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/hexutil"
	"github.com/NilFoundation/nil/nil/internal/abi"
	"github.com/NilFoundation/nil/nil/internal/contracts"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	"github.com/NilFoundation/nil/nil/tests"
	"github.com/stretchr/testify/suite"
)

type SuiteConsensus struct {
	tests.ShardedSuite

	zeroState           *execution.ZeroStateConfig
	testAddress         types.Address
	smartAccountAddress types.Address
	abiTest             *abi.ABI
}

func (s *SuiteConsensus) SetupSuite() {
	var err error
	s.testAddress, err = contracts.CalculateAddress(contracts.NameTest, 1, []byte{1})
	s.Require().NoError(err)

	s.smartAccountAddress = types.MainSmartAccountAddress

	zerostateTmpl := `
contracts:
- name: MainSmartAccount
  address: {{ .SmartAccountAddress }}
  value: 100000000000000000000
  contract: SmartAccount
  ctorArgs: [{{ .MainPublicKey }}]
- name: Test
  address: {{ .TestAddress }}
  value: 100000000000000000000
  contract: tests/Test
`
	zerostateCfg, err := common.ParseTemplate(zerostateTmpl, map[string]any{
		"SmartAccountAddress": s.smartAccountAddress.Hex(),
		"MainPublicKey":       hexutil.Encode(execution.MainPublicKey),
		"TestAddress":         s.testAddress.Hex(),
	})
	s.Require().NoError(err)

	s.zeroState, err = execution.ParseZeroStateConfig(zerostateCfg)
	s.Require().NoError(err)
	s.zeroState.MainPublicKey = execution.MainPublicKey

	s.abiTest, err = contracts.GetAbi(contracts.NameTest)
	s.Require().NoError(err)
}

func (s *SuiteConsensus) SetupTest() {
	nShards := uint32(3)

	s.StartShardAllValidators(&nilservice.Config{
		NShards:              nShards,
		CollatorTickPeriodMs: 200,
		ZeroState:            s.zeroState,
	}, 10625)
}

func (s *SuiteConsensus) TearDownTest() {
	s.Cancel()
}

func (s *SuiteConsensus) TestConsensus() {
	// Check block id grows
	for _, instance := range s.Instances {
		for _, shardId := range instance.Config.MyShards {
			block, err := instance.Client.GetBlock(s.Context, types.ShardId(shardId), "latest", true)
			s.Require().NoError(err)

			s.Require().Eventually(func() bool {
				newBlock, err := instance.Client.GetBlock(s.Context, types.ShardId(shardId), "latest", true)
				s.Require().NoError(err)
				return newBlock != nil && newBlock.Number > block.Number
			}, 30*time.Second, 1*time.Second)
		}
	}

	// Call smart contract
	data, err := s.abiTest.Pack("saveTime")
	s.Require().NoError(err)
	receipt := tests.SendTransactionViaSmartAccount(
		s.T(), s.Instances[0].Client, s.smartAccountAddress, s.testAddress, execution.MainPrivateKey, data)
	s.Require().True(receipt.AllSuccess())

	// Check that all validators have the same block
	nShards := s.GetNShards()
	blocks := make([]*jsonrpc.RPCBlock, 0, nShards)
	for id := range nShards {
		block, err := s.Instances[0].Client.GetBlock(s.Context, types.ShardId(id), "latest", true)
		s.Require().NoError(err)
		blocks = append(blocks, block)
	}

	for _, block := range blocks {
		for _, instance := range s.Instances {
			instanceBlock, err := instance.Client.GetBlock(s.Context, block.ShardId, uint64(block.Number), false)
			s.Require().NoError(err)
			s.Require().NotNil(instanceBlock)
			s.Equal(block.Hash, instanceBlock.Hash)
		}
	}
}

func TestConsensus(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(SuiteConsensus))
}
