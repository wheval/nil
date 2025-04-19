package rollup

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var (
	minBlobGasPrice            = big.NewInt(params.BlobTxMinBlobGasprice)
	blobGaspriceUpdateFraction = big.NewInt(int64(params.DefaultCancunBlobConfig.UpdateFraction))
)

// CalcBlobFee calculates the blobfee from the header's excess blob gas field.
func CalcBlobFee(excessBlobGas uint64) *big.Int {
	return fakeExponential(minBlobGasPrice, new(big.Int).SetUint64(excessBlobGas), blobGaspriceUpdateFraction)
}

// fakeExponential approximates factor * e ** (numerator / denominator) using
// Taylor expansion.
func fakeExponential(factor, numerator, denominator *big.Int) *big.Int {
	var (
		output = new(big.Int)
		accum  = new(big.Int).Mul(factor, denominator)
	)
	for i := 1; accum.Sign() > 0; i++ {
		output.Add(output, accum)

		accum.Mul(accum, numerator)
		accum.Div(accum, denominator)
		accum.Div(accum, big.NewInt(int64(i)))
	}
	return output.Div(output, denominator)
}

func GetBlobGasPrice(header *types.Header) (*uint256.Int, error) {
	if header.ExcessBlobGas == nil {
		return nil, errors.New("GetBlobGasPrice: header is missing excessBlobGas")
	}
	v := CalcBlobFee(*header.ExcessBlobGas)
	return uint256.MustFromBig(v), nil
}
