package execution

import (
	"testing"

	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/stretchr/testify/require"
)

func TestPriceCalculation(t *testing.T) {
	t.Parallel()

	feeCalc := MainFeeCalculator{}

	prevBlock := &types.Block{}

	gasTarget := GasTarget(types.DefaultMaxGasInBlock)
	prevBlock.BaseFee = types.DefaultGasPrice
	prevBlock.GasUsed = gasTarget
	f := feeCalc.CalculateBaseFee(prevBlock)
	require.Equal(t, prevBlock.BaseFee.Uint64(), f.Uint64())

	prevBlock.BaseFee = f
	prevBlock.GasUsed = gasTarget * 2
	f = feeCalc.CalculateBaseFee(prevBlock)
	require.Greater(t, f.Uint64(), prevBlock.BaseFee.Uint64())

	prevBlock.BaseFee = f
	prevBlock.GasUsed = gasTarget - 1_000_000
	f = feeCalc.CalculateBaseFee(prevBlock)
	require.Less(t, f.Uint64(), prevBlock.BaseFee.Uint64())
}
