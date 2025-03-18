package execution

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/abi"
	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/contracts"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/tools/solc"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

type SuiteZeroState struct {
	suite.Suite

	ctx context.Context

	faucetAddr types.Address
	faucetABI  *abi.ABI

	state     *ExecutionState
	contracts map[string]*compiler.Contract
}

func (suite *SuiteZeroState) SetupSuite() {
	var err error
	suite.ctx = context.Background()

	defaultZeroStateConfig, err := CreateDefaultZeroStateConfig(MainPublicKey)
	suite.Require().NoError(err)

	faucetAddress := defaultZeroStateConfig.GetContractAddress("Faucet")
	suite.Require().NotNil(faucetAddress)
	suite.faucetAddr = faucetAddress

	suite.faucetABI, err = contracts.GetAbi(contracts.NameFaucet)
	suite.Require().NoError(err)
}

func (suite *SuiteZeroState) SetupTest() {
	var err error
	suite.state = newState(suite.T())

	suite.contracts, err = solc.CompileSource("./testdata/call.sol")
	suite.Require().NoError(err)
}

func (suite *SuiteZeroState) TearDownTest() {
	suite.state.tx.Rollback()
}

func (suite *SuiteZeroState) getBalance(address types.Address) types.Value {
	suite.T().Helper()

	account, ok := suite.state.Accounts[address]
	suite.Require().True(ok)
	return account.Balance
}

func (suite *SuiteZeroState) TestYamlSerialization() {
	orig, err := CreateDefaultZeroStateConfig(MainPublicKey)
	suite.Require().NoError(err)

	yamlData, err := yaml.Marshal(orig)
	suite.Require().NoError(err)

	deserialized := &ZeroStateConfig{}
	err = yaml.Unmarshal(yamlData, deserialized)
	suite.Require().NoError(err)

	suite.Require().Equal(orig, deserialized)
}

func (suite *SuiteZeroState) TestWithdrawFromFaucet() {
	receiverContract := suite.contracts["SimpleContract"]
	receiverAddr := deployContract(suite.T(), receiverContract, suite.state, 2)
	faucetBalance := suite.getBalance(suite.faucetAddr)

	calldata, err := suite.faucetABI.Pack("withdrawTo", receiverAddr, big.NewInt(100))
	suite.Require().NoError(err)

	gasLimit := types.Gas(100_000).ToValue(types.DefaultGasPrice)
	callTransaction := &types.Transaction{
		TransactionDigest: types.TransactionDigest{
			Data:         calldata,
			To:           suite.faucetAddr,
			FeeCredit:    gasLimit,
			MaxFeePerGas: types.MaxFeePerGasDefault,
		},
		From: suite.faucetAddr,
	}
	res := suite.state.HandleTransaction(suite.ctx, callTransaction, dummyPayer{})
	suite.Require().False(res.Failed())

	outTxnHash, ok := reflect.ValueOf(suite.state.OutTransactions).MapKeys()[0].Interface().(common.Hash)
	suite.Require().True(ok)
	outTxn := suite.state.OutTransactions[outTxnHash][0]
	suite.Require().NotNil(outTxn)

	res = suite.state.HandleTransaction(suite.ctx, outTxn.Transaction, dummyPayer{})
	suite.Require().False(res.Failed())

	faucetBalance = faucetBalance.Sub64(100)
	newFaucetBalance := suite.getBalance(suite.faucetAddr)
	suite.Require().Negative(newFaucetBalance.Cmp(faucetBalance))
	suite.Require().EqualValues(types.NewValueFromUint64(100), suite.getBalance(receiverAddr))
}

func TestZerostateFromConfig(t *testing.T) {
	t.Parallel()

	var state *ExecutionState

	database, err := db.NewBadgerDbInMemory()
	require.NoError(t, err)
	tx, err := database.CreateRwTx(t.Context())
	require.NoError(t, err)
	defer tx.Rollback()

	configAccessor, err := config.NewConfigAccessorTx(tx, nil)
	require.NoError(t, err)
	state, err = NewExecutionState(tx, 0, StateParams{ConfigAccessor: configAccessor})
	require.NoError(t, err)

	zeroState := &ZeroStateConfig{
		ConfigParams: ConfigParams{
			GasPrice: config.ParamGasPrice{
				Shards: []types.Uint256{*types.NewUint256(1), *types.NewUint256(2), *types.NewUint256(3)},
			},
		},
	}
	err = state.GenerateZeroState(zeroState)
	require.NoError(t, err)
	require.Equal(t, 0, state.GasPrice.Cmp(types.NewValueFromUint64(1)))

	state, err = NewExecutionState(tx, 1, StateParams{
		ConfigAccessor: config.GetStubAccessor(),
	})
	require.NoError(t, err)
	zeroState = &ZeroStateConfig{
		ConfigParams: ConfigParams{
			GasPrice: config.ParamGasPrice{
				Shards: []types.Uint256{*types.NewUint256(1), *types.NewUint256(2), *types.NewUint256(3)},
			},
		},
	}

	err = state.GenerateZeroState(zeroState)
	require.NoError(t, err)
	require.Equal(t, 0, state.GasPrice.Cmp(types.NewValueFromUint64(2)))

	state, err = NewExecutionState(tx, 2, StateParams{
		ConfigAccessor: config.GetStubAccessor(),
	})
	require.NoError(t, err)
	zeroState = &ZeroStateConfig{
		ConfigParams: ConfigParams{
			GasPrice: config.ParamGasPrice{
				Shards: []types.Uint256{*types.NewUint256(1), *types.NewUint256(2), *types.NewUint256(3)},
			},
		},
	}

	err = state.GenerateZeroState(zeroState)
	require.NoError(t, err)
	require.Equal(t, 0, state.GasPrice.Cmp(types.NewValueFromUint64(3)))

	smartAccountAddr := types.ShardAndHexToAddress(types.MainShardId, "0x111111111111111111111111111111111111")

	state, err = NewExecutionState(tx, types.MainShardId, StateParams{ConfigAccessor: configAccessor})
	require.NoError(t, err)
	zeroState = &ZeroStateConfig{
		Contracts: []*ContractDescr{
			{Name: "Faucet", Value: types.NewValueFromUint64(87654321), Contract: "Faucet"},
			{
				Name:     "MainSmartAccount",
				Contract: "SmartAccount",
				Address:  smartAccountAddr,
				Value:    types.NewValueFromUint64(12345678),
				CtorArgs: []any{MainPublicKey},
			},
		},
	}
	err = state.GenerateZeroState(zeroState)
	require.NoError(t, err)
	require.Equal(t, types.DefaultGasPrice, state.GasPrice)

	smartAccount, err := state.GetAccount(smartAccountAddr)
	require.NoError(t, err)
	require.NotNil(t, smartAccount)
	require.Equal(t, smartAccount.Balance, types.NewValueFromUint64(12345678))

	faucetCode, err := contracts.GetCode(contracts.NameFaucet)
	require.NoError(t, err)
	faucetAddr := types.CreateAddress(types.MainShardId, types.BuildDeployPayload(faucetCode, common.EmptyHash))

	faucet, err := state.GetAccount(faucetAddr)
	require.NoError(t, err)
	require.NotNil(t, faucet)
	require.Equal(t, faucet.Balance, types.NewValueFromUint64(87654321))

	// Test should fail because contract hasn't `code` item
	state, err = NewExecutionState(tx, types.BaseShardId, StateParams{
		ConfigAccessor: config.GetStubAccessor(),
	})
	require.NoError(t, err)
	zeroState = &ZeroStateConfig{
		Contracts: []*ContractDescr{
			{Name: "Faucet"},
		},
	}
	err = state.GenerateZeroState(zeroState)
	require.Error(t, err)

	// Test only one contract should deployed in specific shard
	state, err = NewExecutionState(tx, types.BaseShardId, StateParams{
		ConfigAccessor: config.GetStubAccessor(),
	})
	require.NoError(t, err)
	zeroState = &ZeroStateConfig{
		Contracts: []*ContractDescr{
			{Name: "Faucet", Value: types.NewValueFromUint64(87654321), Contract: "Faucet", Shard: 1},
			{
				Name: "MainSmartAccount", Contract: "SmartAccount", Address: smartAccountAddr,
				Value: types.NewValueFromUint64(12345678), CtorArgs: []any{MainPublicKey},
			},
		},
	}

	err = state.GenerateZeroState(zeroState)
	require.NoError(t, err)

	faucetAddr = types.CreateAddress(types.BaseShardId, types.BuildDeployPayload(faucetCode, common.EmptyHash))

	faucet, err = state.GetAccount(faucetAddr)
	require.NoError(t, err)
	require.NotNil(t, faucet)
	smartAccount, err = state.GetAccount(smartAccountAddr)
	require.NoError(t, err)
	require.Nil(t, smartAccount)
}

func TestSuiteZeroState(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(SuiteZeroState))
}
