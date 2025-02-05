package config

import (
	"errors"

	"github.com/NilFoundation/nil/nil/common/hexutil"
	"github.com/NilFoundation/nil/nil/internal/types"
)

const ValidatorPubkeySize = 33

type Pubkey [ValidatorPubkeySize]byte

func (k Pubkey) MarshalText() ([]byte, error) {
	return hexutil.Bytes(k[:]).MarshalText()
}

func (k *Pubkey) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("Pubkey", input, k[:])
}

var ParamsList = []IConfigParam{
	new(ParamValidators),
	new(ParamGasPrice),
}

// This is a workaround for fastssz bug where it doesn't add import of `types` package to generated code.
// Adding this struct solves the issue. It can be removed once something other from `types` package will be used in the
// following structs.
type WorkaroundToImportTypes struct {
	Tmp types.TransactionIndex
}

var ErrParamCastFailed = errors.New("input object cannot be cast to Param")

type ListValidators struct {
	List []ValidatorInfo `json:"list" ssz-max:"4096" yaml:"list"`
}

type ParamValidators struct {
	Validators []ListValidators `json:"validators" ssz-max:"4096" yaml:"validators"`
}

type ValidatorInfo struct {
	PublicKey         Pubkey        `json:"pubKey" yaml:"pubKey" ssz-size:"33"`
	WithdrawalAddress types.Address `json:"withdrawalAddress" yaml:"withdrawalAddress"`
}

var _ IConfigParam = new(ParamValidators)

func (p *ParamValidators) Name() string {
	return "curr_validators"
}

func (p *ParamValidators) Accessor() *ParamAccessor {
	return CreateAccessor[ParamValidators]()
}

type ParamGasPrice struct {
	GasPriceScale types.Uint256   `json:"gasPriceScale" yaml:"gasPriceScale"`
	Shards        []types.Uint256 `json:"shards" ssz-max:"4096" yaml:"shards"`
}

var _ IConfigParam = new(ParamGasPrice)

func (p *ParamGasPrice) Name() string {
	return "gas_price"
}

func (p *ParamGasPrice) Accessor() *ParamAccessor {
	return CreateAccessor[ParamGasPrice]()
}

func CreateAccessor[T any, paramPtr IConfigParamPointer[T]]() *ParamAccessor {
	return &ParamAccessor{
		func(c ConfigAccessor) (any, error) {
			return getParamImpl[T, paramPtr](c)
		},
		func(c ConfigAccessor, v any) error {
			if param, ok := v.(*T); ok {
				return setParamImpl[T](c, param)
			}
			return ErrParamCastFailed
		},
		func(v any) ([]byte, error) {
			if param, ok := v.(*T); ok {
				return packSolidityImpl[T](param)
			}
			return nil, ErrParamCastFailed
		},
		func(data []byte) (any, error) { return unpackSolidityImpl[T](data) },
	}
}

func GetParamValidators(c ConfigAccessor) (*ParamValidators, error) {
	return getParamImpl[ParamValidators](c)
}

func SetParamValidators(c ConfigAccessor, params *ParamValidators) error {
	return setParamImpl(c, params)
}

func GetParamGasPrice(c ConfigAccessor) (*ParamGasPrice, error) {
	return getParamImpl[ParamGasPrice](c)
}

func SetParamGasPrice(c ConfigAccessor, params *ParamGasPrice) error {
	return setParamImpl(c, params)
}

func GetParamNShards(c ConfigAccessor) (uint32, error) {
	param, err := getParamImpl[ParamGasPrice](c)
	if err != nil {
		return 0, err
	}
	return uint32(len(param.Shards)), nil
}
