package execution

import (
	"fmt"

	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/shopspring/decimal"
)

func (es *ExecutionState) UpdateBaseFee() error {
	acc, err := es.shardAccessor.GetBlock().ByHash(es.PrevBlock)
	if err != nil {
		// If we can't read the previous block, we don't change the gas price
		es.GasPrice = types.DefaultGasPrice
		logger.Error().Err(err).Msg("failed to read previous block, gas price won't be changed")
		return fmt.Errorf("failed to read previous block: %w", err)
	}
	prevBlock := acc.Block()

	es.BaseFee = calculateBaseFee(prevBlock.BaseFee, prevBlock.GasUsed)

	if es.BaseFee.Cmp(prevBlock.BaseFee) != 0 {
		logger.Debug().
			Stringer("Old", prevBlock.BaseFee).
			Stringer("New", es.BaseFee).
			Msg("Gas price updated")
	}

	return nil
}

func GasTarget(gasLimit types.Gas) types.Gas {
	return gasLimit / 2
}

var (
	// Multiplier for base fee adjustment
	adjustmentFactor = decimal.New(1, 0).Div(decimal.New(1013400, -4))
	// Gas limit normalized to percentage (100% = 100.0)
	gasLimitPercentage = decimal.New(1000000, -4)
	// Smoothing factor for sigmoids
	smoothingFactor = decimal.New(50000, -4)
	// Center of first sigmoid (25% of gas limit)
	centerSigmoid1 = decimal.New(2500, -4).Mul(gasLimitPercentage)
	// Center of second sigmoid (75% of gas limit)
	centerSigmoid2 = decimal.New(7500, -4).Mul(gasLimitPercentage)
	blockGasLimit  = decimal.NewFromInt(int64(types.DefaultGasLimit.Uint64()))
	maxSigmaDiff   decimal.Decimal
	eNumber        = decimal.New(27182818284, -10)
)

func init() {
	a := sigmoid2(gasLimitPercentage, centerSigmoid2, smoothingFactor).
		Sub(sigmoid1(gasLimitPercentage, centerSigmoid1, smoothingFactor))
	b := sigmoid2(decimal.New(0, -4), centerSigmoid2, smoothingFactor).
		Sub(sigmoid1(decimal.New(0, -4), centerSigmoid1, smoothingFactor))
	if a.Cmp(b) >= 0 {
		maxSigmaDiff = a
	} else {
		maxSigmaDiff = b
	}
}

func calculateBaseFee(baseFeePrevious types.Value, gasUsedPrevious types.Gas) types.Value {
	// Convert to percentage
	gasUsedPercentage := toPercentage(decimal.NewFromInt(int64(gasUsedPrevious.Uint64())), blockGasLimit)

	// Calculate the difference between the two sigmoids
	sigmaDiff := sigmoid2(gasUsedPercentage, centerSigmoid2, smoothingFactor).Sub(
		sigmoid1(gasUsedPercentage, centerSigmoid1, smoothingFactor))

	// Normalize the difference
	normalized := sigmaDiff.Div(maxSigmaDiff)

	baseFeePreviousFixed := decimal.NewFromBigInt(baseFeePrevious.Int().ToBig(), 0)
	newFee := baseFeePreviousFixed.Mul(adjustmentFactor.Mul(normalized).Add(decimal.New(1, 0)))

	if newFee.Cmp(decimal.NewFromBigInt(types.DefaultGasPrice.ToBig(), 0)) > 0 {
		return types.NewValueFromBigMust(newFee.BigInt())
	}

	return types.DefaultGasPrice
}

func toPercentage(gasUsedPrevious, gasLimitAbsolute decimal.Decimal) decimal.Decimal {
	return gasUsedPrevious.Mul(decimal.New(100, 0)).Div(gasLimitAbsolute)
}

func sigmoid1(gasUsedPercentage, centerSigmoid, smoothingFactor decimal.Decimal) decimal.Decimal {
	n := gasUsedPercentage.Sub(centerSigmoid).Div(smoothingFactor)
	exp := eNumber.Pow(n)
	return decimal.NewFromInt(1).Div(decimal.NewFromInt(1).Add(exp))
}

func sigmoid2(gasUsedPercentage, centerSigmoid, smoothingFactor decimal.Decimal) decimal.Decimal {
	n := gasUsedPercentage.Sub(centerSigmoid).Neg().Div(smoothingFactor)
	exp := eNumber.Pow(n)
	return decimal.NewFromInt(1).Div(decimal.NewFromInt(1).Add(exp))
}
