package rollup

import (
	"errors"

	"github.com/ethereum/go-ethereum/consensus/misc/eip4844"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

func GetBlobGasPrice(header *types.Header) (*uint256.Int, error) {
	if header.ExcessBlobGas == nil {
		return nil, errors.New("GetBlobGasPrice: header is missing excessBlobGas")
	}
	v := eip4844.CalcBlobFee(*header.ExcessBlobGas)
	return uint256.MustFromBig(v), nil
}
