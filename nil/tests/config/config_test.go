package main

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/NilFoundation/nil/nil/internal/abi"
	"github.com/NilFoundation/nil/nil/internal/collate"
	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/contracts"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/keys"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/services/rpc/jsonrpc"
	"github.com/NilFoundation/nil/nil/tests"
	"github.com/stretchr/testify/suite"
)

type SuiteConfigParams struct {
	tests.RpcSuite
	testAddressMain   types.Address
	testAddress       types.Address
	abiTest           *abi.ABI
	abiSmartAccount   *abi.ABI
	validatorsKeyPath string
	validatorInfo     config.ValidatorInfo
}

func (s *SuiteConfigParams) SetupSuite() {
	s.ShardsNum = 4

	var err error
	s.testAddressMain, err = contracts.CalculateAddress(contracts.NameConfigTest, types.MainShardId, nil)
	s.Require().NoError(err)

	s.testAddress, err = contracts.CalculateAddress(contracts.NameConfigTest, types.BaseShardId, nil)
	s.Require().NoError(err)

	s.abiSmartAccount, err = contracts.GetAbi("SmartAccount")
	s.Require().NoError(err)

	s.abiTest, err = contracts.GetAbi(contracts.NameConfigTest)
	s.Require().NoError(err)

	s.validatorsKeyPath = s.T().TempDir() + "/validator-keys.yaml"
	km := keys.NewValidatorKeyManager(s.validatorsKeyPath)
	s.Require().NotNil(km)
	s.Require().NoError(km.InitKey())

	pk, err := km.GetPublicKey()
	s.Require().NoError(err)

	s.validatorInfo = config.ValidatorInfo{
		PublicKey: config.Pubkey(pk),
	}
}

func (s *SuiteConfigParams) SetupTest() {
	var err error
	s.Db, err = db.NewBadgerDbInMemory()
	s.Require().NoError(err)
	s.Context, s.CtxCancel = context.WithCancel(context.Background())
}

func (s *SuiteConfigParams) TearDownTest() {
	s.Cancel()
}

func (s *SuiteConfigParams) NewValidator() *config.ValidatorInfo {
	s.T().Helper()

	var pubkey config.Pubkey
	_, err := rand.Read(pubkey[:])
	s.Require().NoError(err)

	address := make([]byte, types.AddrSize)
	_, err = rand.Read(address)
	s.Require().NoError(err)

	return &config.ValidatorInfo{
		PublicKey:         pubkey,
		WithdrawalAddress: types.BytesToAddress(address),
	}
}

func (s *SuiteConfigParams) makeParamValidators(vals ...config.ValidatorInfo) config.ParamValidators {
	s.T().Helper()

	validators := make([]config.ListValidators, 0, s.ShardsNum)
	for range s.ShardsNum {
		validators = append(validators, config.ListValidators{List: vals})
	}
	return config.ParamValidators{Validators: validators}
}

// TODO(@isergeyam): add read/write validators test

func (s *SuiteConfigParams) TestConfigReadWriteGasPrice() {
	cfg := execution.ZeroStateConfig{
		ConfigParams: execution.ConfigParams{
			GasPrice:   config.ParamGasPrice{GasPriceScale: *types.NewUint256(10)},
			Validators: s.makeParamValidators(s.validatorInfo),
		},
		Contracts: []*execution.ContractDescr{
			{
				Name:     "TestConfig",
				Address:  &s.testAddressMain,
				Value:    types.GasToValue(10_000_000_000),
				Contract: contracts.NameConfigTest,
			},
			{
				Name:     "TestConfig",
				Address:  &s.testAddress,
				Value:    types.GasToValue(10_000_000_000),
				Contract: contracts.NameConfigTest,
			},
		},
	}

	// Manually set gas price for all shards. It is necessary because the initial prices are set only during the first
	// block generation. But we will likely read config before that.
	cfg.ConfigParams.GasPrice.Shards = make([]types.Uint256, s.ShardsNum)
	for i := range s.ShardsNum {
		cfg.ConfigParams.GasPrice.Shards[i] = *types.DefaultGasPrice.Uint256
	}

	s.Start(&nilservice.Config{
		NShards:              s.ShardsNum,
		Topology:             collate.TrivialShardTopologyId,
		ZeroState:            &cfg,
		CollatorTickPeriodMs: 100,
		RunMode:              nilservice.CollatorsOnlyRunMode,
		ValidatorKeysPath:    s.validatorsKeyPath,
	})

	var (
		receipt *jsonrpc.RPCReceipt
		data    []byte
	)

	gasPrice := s.readGasPrices()
	gasPrice.GasPriceScale = *types.NewUint256(10)

	s.Run("Check initial gas price param", func() {
		data = s.AbiPack(s.abiTest, "testParamGasPriceEqual", gasPrice)
		receipt = s.SendExternalTransactionNoCheck(data, s.testAddress)
		s.Require().True(receipt.AllSuccess())
	})

	s.Run("Modify param", func() {
		gasPrice.GasPriceScale = *types.NewUint256(123)
		data = s.AbiPack(s.abiTest, "setParamGasPrice", gasPrice)
		receipt = s.SendExternalTransactionNoCheck(data, s.testAddressMain)
		s.Require().True(receipt.AllSuccess())

		data = s.AbiPack(s.abiTest, "testParamGasPriceEqual", gasPrice)
		receipt = s.SendExternalTransactionNoCheck(data, s.testAddressMain)
		s.Require().True(receipt.AllSuccess())

		realGasPrice := s.readGasPrices()

		s.Require().Equal(gasPrice.GasPriceScale, realGasPrice.GasPriceScale)
	})

	s.Run("Read param after write", func() {
		data = s.AbiPack(s.abiTest, "readParamAfterWrite")
		receipt = s.SendExternalTransactionNoCheck(data, s.testAddressMain)
		s.Require().True(receipt.AllSuccess())
	})
}

func (s *SuiteConfigParams) readGasPrices() *config.ParamGasPrice {
	s.T().Helper()

	tx, err := s.Db.CreateRoTx(s.Context)
	s.Require().NoError(err)
	defer tx.Rollback()
	cfgReader, err := config.NewConfigReader(tx, nil)
	s.Require().NoError(err)
	gasPrice, err := config.GetParamGasPrice(cfgReader)
	s.Require().NoError(err)
	return gasPrice
}

func TestConfig(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(SuiteConfigParams))
}
