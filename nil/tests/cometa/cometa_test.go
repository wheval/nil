package cometa

import (
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/assert"
	"github.com/NilFoundation/nil/nil/common/hexutil"
	"github.com/NilFoundation/nil/nil/internal/contracts"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/cometa"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/services/rpc"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	"github.com/NilFoundation/nil/nil/tests"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/suite"
)

type SuiteCometa struct {
	tests.RpcSuite
	cometaClient cometa.Client
	cometaCfg    cometa.Config
	testAddress  types.Address
}

type SuiteCometaBadger struct {
	SuiteCometa
}

type SuiteCometaClickhouse struct {
	SuiteCometa
	clickhouse *exec.Cmd
}

func (s *SuiteCometaClickhouse) SetupSuite() {
	s.cometaCfg.UseBadger = false

	s.cometaCfg.ResetToDefault()
	s.cometaCfg.DbEndpoint = "127.0.0.1:9002"

	suiteSetupDone := false

	defer func() {
		if !suiteSetupDone {
			s.TearDownSuite()
		}
	}()

	dir := s.T().TempDir()
	s.clickhouse = exec.Command( //nolint:gosec
		"clickhouse", "server", "--",
		"--listen_host=0.0.0.0",
		"--tcp_port=9002",
		"--http_port=",
		"--mysql_port=",
		"--path="+dir,
	)
	s.clickhouse.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	s.clickhouse.Dir = dir
	err := s.clickhouse.Start()
	s.Require().NoError(err)

	time.Sleep(1 * time.Second)

	s.SuiteCometa.SetupSuite()

	suiteSetupDone = true
}

func (s *SuiteCometaClickhouse) TearDownSuite() {
	if s.clickhouse != nil {
		// https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
		// simple s.clickhouse.Kill() won't work on child process
		// this leads to errors in sequential test runs
		pgid, err := syscall.Getpgid(s.clickhouse.Process.Pid)
		s.Require().NoError(err)
		err = syscall.Kill(-pgid, syscall.SIGTERM)
		s.Require().NoError(err)
	}
}

func (s *SuiteCometaBadger) SetupSuite() {
	s.cometaCfg.ResetToDefault()
	s.cometaCfg.UseBadger = true
	s.SuiteCometa.SetupSuite()
}

func (s *SuiteCometa) SetupSuite() {
	s.cometaCfg.DbPath = s.T().TempDir() + "/cometa.db"
	s.cometaCfg.OwnEndpoint = ""
	var err error

	s.testAddress, err = contracts.CalculateAddress(contracts.NameTest, 1, []byte{1})
	s.Require().NoError(err)

	mainSmartAccountValue, err := types.NewValueFromDecimal("1000000000000000000000")
	s.Require().NoError(err)
	zerostateCfg := &execution.ZeroStateConfig{
		Contracts: []*execution.ContractDescr{
			{
				Name:     "MainSmartAccount",
				Contract: "SmartAccount",
				Address:  types.MainSmartAccountAddress,
				Value:    mainSmartAccountValue,
				CtorArgs: []any{execution.MainPublicKey},
			},
			{
				Name:     "Tests",
				Contract: "tests/Test",
				Address:  s.testAddress,
				Value:    types.NewValueFromUint64(100000000),
			},
		},
	}
	s.cometaCfg.DbPath = s.T().TempDir() + "/cometa.db"
	s.Start(&nilservice.Config{
		NShards:              2,
		CollatorTickPeriodMs: 200,
		HttpUrl:              rpc.GetSockPath(s.T()),
		Cometa:               &s.cometaCfg,
		ZeroState:            zerostateCfg,
		DisableConsensus:     true,
	})
	s.cometaClient = *cometa.NewClient(s.Endpoint)
	tests.WaitShardTick(s.T(), s.Client, types.MainShardId)
	tests.WaitShardTick(s.T(), s.Client, types.BaseShardId)
}

func (s *SuiteCometa) TestTwinContracts() {
	pk, err := crypto.GenerateKey()
	s.Require().NoError(err)
	pub := crypto.CompressPubkey(&pk.PublicKey)
	smartAccountCode := contracts.PrepareDefaultSmartAccountForOwnerCode(pub)
	deployCode1 := types.BuildDeployPayload(smartAccountCode, common.EmptyHash)
	deployCode2 := types.BuildDeployPayload(smartAccountCode, common.HexToHash("0x1234"))

	smartAccountAddr1, _ := s.DeployContractViaMainSmartAccount(
		types.BaseShardId, deployCode1, s.GasToValue(10_000_000))
	smartAccountAddr2, _ := s.DeployContractViaMainSmartAccount(
		types.BaseShardId, deployCode2, s.GasToValue(10_000_000))

	err = s.cometaClient.RegisterContractFromFile(
		s.Context, "../../contracts/solidity/compile-smart-account.json", smartAccountAddr1)
	s.Require().NoError(err)

	contract1, err := s.cometaClient.GetContractFields(s.Context, smartAccountAddr1, []string{"Name", "InitCode"})
	s.Require().NoError(err)

	contract2, err := s.cometaClient.GetContractFields(s.Context, smartAccountAddr2, []string{"Name", "InitCode"})
	s.Require().NoError(err)

	s.Require().Equal(contract1, contract2)
}

func (s *SuiteCometa) TestGeneratedCode() {
	if !s.cometaCfg.UseBadger {
		s.T().Skip()
	}
	var (
		receipt *jsonrpc.RPCReceipt
		data    []byte
		loc     *cometa.Location
	)
	testAbi, err := contracts.GetAbi(contracts.NameTest)
	s.Require().NoError(err)

	contractData, err := s.cometaClient.CompileContract(s.Context, "../../contracts/solidity/tests/compile-test.json")
	s.Require().NoError(err)
	deployCode := types.BuildDeployPayload(contractData.InitCode, common.EmptyHash)
	testAddress, _ := s.DeployContractViaMainSmartAccount(types.BaseShardId, deployCode, s.GasToValue(10_000_000))

	err = s.cometaClient.RegisterContractData(s.Context, contractData, testAddress)
	s.Require().NoError(err)

	data = []byte("invalid calldata")
	receipt = s.SendExternalTransactionNoCheck(data, testAddress)
	s.Require().False(receipt.AllSuccess())

	loc, err = s.cometaClient.GetLocation(s.Context, testAddress, uint64(receipt.FailedPc))
	s.Require().NoError(err)
	s.Require().Equal("Test.sol:8, function: #function_selector", loc.String())

	data = s.AbiPack(testAbi, "makeFail", int32(1))
	receipt = s.SendExternalTransactionNoCheck(data, testAddress)
	s.Require().False(receipt.AllSuccess())
	s.Require().NotZero(receipt.FailedPc)

	loc, err = s.cometaClient.GetLocation(s.Context, testAddress, uint64(receipt.FailedPc))
	s.Require().NoError(err)
	s.Require().Equal(
		"#utility.yul:8, function: revert_error_dbdddcbe895c83990c08b3492a0e83918d802a52331272ac6fdb6a7c4aea3b1b",
		loc.String())
}

func (s *SuiteCometa) TestMethodList() {
	testAbi, err := contracts.GetAbi(contracts.NameTest)
	s.Require().NoError(err)
	smartAccountAbi, err := contracts.GetAbi(contracts.NameSmartAccount)
	s.Require().NoError(err)

	err = s.cometaClient.RegisterContractFromFile(
		s.Context,
		"../../contracts/solidity/compile-smart-account.json",
		types.MainSmartAccountAddress)
	s.Require().NoError(err)

	err = s.cometaClient.RegisterContractFromFile(
		s.Context, "../../contracts/solidity/tests/compile-test.json", s.testAddress)
	s.Require().NoError(err)

	transactions := []cometa.TransactionInfo{
		{
			Address: s.testAddress,
			FuncId:  hexutil.EncodeNo0x(testAbi.Methods["makeFail"].ID),
		},
		{
			Address: s.testAddress,
			FuncId:  "11111111",
		},
		{
			Address: types.MainSmartAccountAddress,
			FuncId:  hexutil.Encode(smartAccountAbi.Methods["asyncCall"].ID),
		},
	}
	res, err := s.cometaClient.DecodeTransactionsCallData(s.Context, transactions)
	s.Require().NoError(err)
	s.Require().Equal("makeFail(int32)", res[0])
	s.Require().Empty(res[1])
	s.Require().Equal("asyncCall(address,address,address,(address,uint256)[],uint256,bytes)", res[2])

	// Obviously wrong funcId, should return error
	transactions = []cometa.TransactionInfo{
		{
			Address: s.testAddress,
			FuncId:  "123",
		},
	}
	_, err = s.cometaClient.DecodeTransactionsCallData(s.Context, transactions)
	s.Require().Error(err)
}

func checkClickhouseInstalled() bool {
	cmd := exec.Command("clickhouse", "--version")
	err := cmd.Run()
	return err == nil
}

func TestCometaClickhouse(t *testing.T) {
	if !checkClickhouseInstalled() {
		if assert.Enable {
			t.Fatal("Clickhouse is not installed")
		} else {
			t.Skip("Clickhouse is not installed")
		}
	}
	t.Parallel()
	suite.Run(t, new(SuiteCometaClickhouse))
}

func TestCometaBadger(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SuiteCometaBadger))
}
