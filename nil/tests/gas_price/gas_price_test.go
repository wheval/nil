package tests

import (
	"testing"

	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/nilservice"
	"github.com/NilFoundation/nil/nil/services/rpc"
	"github.com/NilFoundation/nil/nil/tests"
	"github.com/stretchr/testify/suite"
)

type SuitGasPrice struct {
	tests.RpcSuite
}

func (s *SuitGasPrice) SetupSuite() {
	s.Start(&nilservice.Config{
		NShards:       4,
		HttpUrl:       rpc.GetSockPath(s.T()),
		ZeroStateYaml: execution.DefaultZeroStateConfig,
		GasPriceScale: 15,
		GasBasePrice:  types.DefaultGasPrice.Uint64(),
		RunMode:       nilservice.CollatorsOnlyRunMode,
	})
}

func (s *SuitGasPrice) TearDownSuite() {
	s.Cancel()
}

func (s *SuitGasPrice) TestGasBehaviour() {
	// TODO: implement
}

func TestSuiteGasPrice(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(SuitGasPrice))
}
