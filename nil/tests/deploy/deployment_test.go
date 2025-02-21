package main

import (
	"math/big"
	"testing"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/hexutil"
	"github.com/NilFoundation/nil/nil/internal/abi"
	"github.com/NilFoundation/nil/nil/internal/contracts"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/services/rpc"
	"github.com/NilFoundation/nil/nil/tests"
	"github.com/stretchr/testify/suite"
)

type SuiteDeployment struct {
	tests.RpcSuite
	addressDeployer types.Address
	abiDeployer     *abi.ABI
	abiDeployee     *abi.ABI
	zerostateCfg    string
}

func (s *SuiteDeployment) SetupSuite() {
	s.ShardsNum = 4

	var err error
	s.addressDeployer, err = contracts.CalculateAddress(contracts.NameDeployer, 1, []byte{1})
	s.Require().NoError(err)

	zerostateTmpl := `
contracts:
- name: MainSmartAccount
  address: {{ .SmartAccountAddress }}
  value: 100000000000000
  contract: SmartAccount
  ctorArgs: [{{ .MainPublicKey }}]
- name: Deployer
  address: {{ .DeployerAddress }}
  value: 100000000000000
  contract: tests/Deployer
`
	s.zerostateCfg, err = common.ParseTemplate(zerostateTmpl, map[string]any{
		"SmartAccountAddress": types.MainSmartAccountAddress.Hex(),
		"MainPublicKey":       hexutil.Encode(execution.MainPublicKey),
		"DeployerAddress":     s.addressDeployer.Hex(),
	})
	s.Require().NoError(err)

	s.abiDeployer, err = contracts.GetAbi(contracts.NameDeployer)
	s.Require().NoError(err)

	s.abiDeployee, err = contracts.GetAbi(contracts.NameDeployee)
	s.Require().NoError(err)
}

func (s *SuiteDeployment) SetupTest() {
	var err error
	zeroState, err := execution.ParseZeroStateConfig(s.zerostateCfg)
	s.Require().NoError(err)
	zeroState.MainPublicKey = execution.MainPublicKey

	s.Start(&nilservice.Config{
		NShards:              s.ShardsNum,
		HttpUrl:              rpc.GetSockPath(s.T()),
		ZeroState:            zeroState,
		CollatorTickPeriodMs: 300,
		RunMode:              nilservice.CollatorsOnlyRunMode,
	})
}

func (s *SuiteDeployment) TearDownTest() {
	s.Cancel()
}

func (s *SuiteDeployment) TestDeploy() {
	s.Run("deploy", func() {
		calldata := s.AbiPack(s.abiDeployer, "deploy", big.NewInt(1), uint32(1234), big.NewInt(5678), big.NewInt(1111))
		receipt := s.SendExternalTransaction(calldata, s.addressDeployer)
		s.Require().True(receipt.AllSuccess())

		res := s.CallGetter(s.addressDeployer, s.AbiPack(s.abiDeployer, "deployee"), "latest", nil)
		address := types.BytesToAddress(res)
		s.Require().Equal(types.ShardId(1), address.ShardId())

		res = s.CallGetter(address, s.AbiPack(s.abiDeployee, "deployer"), "latest", nil)
		s.Require().Equal(s.addressDeployer, types.BytesToAddress(res))

		num := tests.CallGetterT[uint32](s.T(), s.Context, s.Client, s.abiDeployee, address, "num")
		s.Require().Equal(uint32(1234), num)

		s.Require().Equal(types.NewValueFromUint64(1111), s.GetBalance(address))
	})

	s.Run("deploy via smart account", func() {
		salt := big.NewInt(789878)

		abiSmartAccount, err := contracts.GetAbi(contracts.NameSmartAccount)
		s.Require().NoError(err)
		code, err := contracts.GetCode(contracts.NameDeployee)
		s.Require().NoError(err)
		deployPayload := s.PrepareDefaultDeployPayload(*s.abiDeployee, code, s.addressDeployer, uint32(987654321))

		calldata := s.AbiPack(abiSmartAccount, "asyncDeploy", big.NewInt(2), big.NewInt(1111), deployPayload.Bytes(), salt)
		receipt := s.SendExternalTransaction(calldata, types.MainSmartAccountAddress)
		s.Require().True(receipt.AllSuccess())

		address := types.CreateAddress(types.ShardId(2), types.BuildDeployPayload(deployPayload.Bytes(),
			common.BigToHash(salt)))

		res := s.CallGetter(address, s.AbiPack(s.abiDeployee, "deployer"), "latest", nil)
		s.Require().Equal(s.addressDeployer, types.BytesToAddress(res))

		num := tests.CallGetterT[uint32](s.T(), s.Context, s.Client, s.abiDeployee, address, "num")
		s.Require().Equal(uint32(987654321), num)

		s.Require().Equal(types.NewValueFromUint64(1111), s.GetBalance(address))
	})
}

func TestDeployment(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SuiteDeployment))
}
