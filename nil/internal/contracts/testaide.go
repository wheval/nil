//go:build test

package contracts

import (
	"testing"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	NameCounter                    = "tests/Counter"
	NameDeployer                   = "tests/Deployer"
	NameDeployee                   = "tests/Deployee"
	NameTransactionCheck           = "tests/TransactionCheck"
	NameSender                     = "tests/Sender"
	NameTest                       = "tests/Test"
	NameTokensTest                 = "tests/TokensTest"
	NameTokensTestNoExternalAccess = "tests/TokensTestNoExternalAccess"
	NameRequestResponseTest        = "tests/RequestResponseTest"
	NamePrecompilesTest            = "tests/PrecompilesTest"
	NameConfigTest                 = "tests/ConfigTest"
	NameStresser                   = "tests/Stresser"
)

func GetDeployPayload(t *testing.T, name string) types.DeployPayload {
	t.Helper()
	return GetDeployPayloadWithSalt(t, name, common.EmptyHash)
}

func GetDeployPayloadWithSalt(t *testing.T, name string, salt common.Hash) types.DeployPayload {
	t.Helper()

	code, err := GetCode(name)
	require.NoError(t, err)
	return types.BuildDeployPayload(code, salt)
}

func CounterDeployPayload(t *testing.T) types.DeployPayload {
	t.Helper()

	return CounterDeployPayloadWithSalt(t, common.EmptyHash)
}

func CounterDeployPayloadWithSalt(t *testing.T, salt common.Hash) types.DeployPayload {
	t.Helper()

	return GetDeployPayloadWithSalt(t, NameCounter, salt)
}

func CounterAddress(t *testing.T, shardId types.ShardId) types.Address {
	t.Helper()

	return CounterAddressWithSalt(t, shardId, common.EmptyHash)
}

func CounterAddressWithSalt(t *testing.T, shardId types.ShardId, salt common.Hash) types.Address {
	t.Helper()

	return types.CreateAddress(shardId, CounterDeployPayloadWithSalt(t, salt))
}

func FaucetDeployPayload(t *testing.T) types.DeployPayload {
	t.Helper()

	return GetDeployPayload(t, NameFaucet)
}

func SmartAccountAddress(t *testing.T, shardId types.ShardId, salt, pubKey []byte) types.Address {
	t.Helper()

	res, err := CalculateAddress(NameSmartAccount, shardId, salt, pubKey)
	require.NoError(t, err)
	return res
}

func NewCallDataT(t *testing.T, fileName, methodName string, args ...any) []byte {
	t.Helper()

	callData, err := NewCallData(fileName, methodName, args...)
	require.NoError(t, err)

	return callData
}

func NewCounterAddCallData(t *testing.T, value int32) []byte {
	t.Helper()

	return NewCallDataT(t, NameCounter, "add", value)
}

func NewCounterGetCallData(t *testing.T) []byte {
	t.Helper()

	return NewCallDataT(t, NameCounter, "get")
}

func NewFaucetWithdrawToCallData(t *testing.T, dst types.Address, value types.Value) []byte {
	t.Helper()

	return NewCallDataT(t, NameFaucet, "withdrawTo", dst, value.ToBig())
}

func GetCounterValue(t *testing.T, data []byte) int32 {
	t.Helper()

	res, err := UnpackData(NameCounter, "get", data)
	require.NoError(t, err)

	val, ok := res[0].(int32)
	require.True(t, ok)
	return val
}

func NewSmartAccountAsyncCallCallData(t *testing.T, bytecode types.Code,
	value types.Value, tokens []types.TokenBalance, contractAddress types.Address,
) []byte {
	t.Helper()

	return NewCallDataT(t, NameSmartAccount, "asyncCall", contractAddress,
		types.EmptyAddress, types.EmptyAddress, tokens, value, []byte(bytecode))
}
