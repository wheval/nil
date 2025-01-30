package execution

import (
	"fmt"
	"testing"

	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPriceCalculation(t *testing.T) {
	t.Parallel()

	a, _ := decimal.NewFromString("12.3400")
	b := decimal.New(5678, -3)
	c := a.Add(b)
	fmt.Println(c.String())

	baseFeePrevious := types.DefaultGasPrice
	gasTarget := GasTarget(types.DefaultGasLimit)
	gasUsedPrevious := gasTarget
	f := calculateBaseFee(baseFeePrevious, gasUsedPrevious)
	require.Equal(t, baseFeePrevious.Uint64(), f.Uint64())

	gasUsedPrevious *= 2
	baseFeePrevious = f
	f = calculateBaseFee(baseFeePrevious, gasUsedPrevious)
	require.Greater(t, f.Uint64(), baseFeePrevious.Uint64())

	baseFeePrevious = f
	gasUsedPrevious = gasTarget - 100_000
	f = calculateBaseFee(baseFeePrevious, gasUsedPrevious)
	require.Less(t, f.Uint64(), baseFeePrevious.Uint64())
}
