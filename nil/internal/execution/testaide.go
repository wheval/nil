//go:build test

package execution

import (
	"context"
	"testing"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/contracts"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	DefaultGasLimit = 100_000
)

var (
	DefaultGasCredit = types.Gas(DefaultGasLimit).ToValue(types.DefaultGasPrice)

	DefaultSendValue = types.GasToValue(200_000_000)
)

func GenerateZeroState(t *testing.T, ctx context.Context,
	shardId types.ShardId, txFabric db.DB,
) common.Hash {
	t.Helper()

	g, err := NewBlockGenerator(ctx,
		NewBlockGeneratorParams(shardId, 1),
		txFabric, nil, nil)
	require.NoError(t, err)
	defer g.Rollback()

	zerostateCfg, err := ParseZeroStateConfig(DefaultZeroStateConfig)
	require.NoError(t, err)
	zerostateCfg.ConfigParams = ConfigParams{
		GasPrice: config.ParamGasPrice{
			Shards: []types.Uint256{*types.NewUint256(10), *types.NewUint256(10), *types.NewUint256(10)},
		},
	}

	block, err := g.GenerateZeroState("", zerostateCfg)
	require.NoError(t, err)
	require.NotNil(t, block)
	return block.Hash(shardId)
}

func GenerateBlockFromTransactions(t *testing.T, ctx context.Context,
	shardId types.ShardId, blockId types.BlockNumber, prevBlock common.Hash,
	txFabric db.DB, childChainBlocks map[types.ShardId]common.Hash, txns ...*types.Transaction,
) common.Hash {
	t.Helper()
	return generateBlockFromTransactions(t, ctx, true, shardId, blockId, prevBlock, txFabric, childChainBlocks, txns...)
}

func GenerateBlockFromTransactionsWithoutExecution(t *testing.T, ctx context.Context,
	shardId types.ShardId, blockId types.BlockNumber, prevBlock common.Hash,
	txFabric db.DB, txns ...*types.Transaction,
) common.Hash {
	t.Helper()
	return generateBlockFromTransactions(t, ctx, false, shardId, blockId, prevBlock, txFabric, nil, txns...)
}

func generateBlockFromTransactions(t *testing.T, ctx context.Context, execute bool,
	shardId types.ShardId, blockId types.BlockNumber, prevBlock common.Hash,
	txFabric db.DB, childChainBlocks map[types.ShardId]common.Hash, txns ...*types.Transaction,
) common.Hash {
	t.Helper()

	tx, err := txFabric.CreateRwTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback()

	es, err := NewExecutionState(tx, shardId, StateParams{
		BlockHash:      prevBlock,
		Timer:          common.NewTestTimer(0),
		ConfigAccessor: config.GetStubAccessor(),
	})
	require.NoError(t, err)
	es.BaseFee = types.DefaultGasPrice

	for _, txn := range txns {
		es.AddInTransaction(txn)

		if !execute {
			es.AddReceipt(NewExecutionResult())
			continue
		}

		var execResult *ExecutionResult
		if txn.IsRefund() {
			execResult = NewExecutionResult()
			execResult.SetFatal(es.handleRefundTransaction(ctx, txn))
		} else {
			execResult = es.HandleTransaction(ctx, txn, dummyPayer{})
		}
		require.False(t, execResult.Failed())

		es.AddReceipt(execResult)
	}

	es.ChildChainBlocks = childChainBlocks

	blockRes, err := es.Commit(blockId, nil)
	require.NoError(t, err)

	err = PostprocessBlock(tx, shardId, blockRes)
	require.NoError(t, err)
	require.NotNil(t, blockRes.Block)

	require.NoError(t, db.WriteBlockTimestamp(tx, shardId, blockRes.BlockHash, 0))

	require.NoError(t, tx.Commit())

	return blockRes.BlockHash
}

func NewDeployTransaction(payload types.DeployPayload,
	shardId types.ShardId, from types.Address, seqno types.Seqno, value types.Value,
) *types.Transaction {
	return &types.Transaction{
		TransactionDigest: types.TransactionDigest{
			Flags:        types.NewTransactionFlags(types.TransactionFlagInternal, types.TransactionFlagDeploy),
			Data:         payload.Bytes(),
			Seqno:        seqno,
			FeeCredit:    types.GasToValue(10_000_000),
			To:           types.CreateAddress(shardId, payload),
			MaxFeePerGas: types.MaxFeePerGasDefault,
		},
		From:  from,
		Value: value,
	}
}

func NewExecutionTransaction(from, to types.Address, seqno types.Seqno, callData []byte) *types.Transaction {
	return &types.Transaction{
		TransactionDigest: types.TransactionDigest{
			To:           to,
			Data:         callData,
			Seqno:        seqno,
			FeeCredit:    DefaultGasCredit,
			MaxFeePerGas: types.MaxFeePerGasDefault,
		},
		From: from,
	}
}

func NewSendMoneyTransaction(t *testing.T, to types.Address, seqno types.Seqno) *types.Transaction {
	t.Helper()

	m := NewExecutionTransaction(types.MainSmartAccountAddress, types.MainSmartAccountAddress, seqno,
		contracts.NewSmartAccountSendCallData(t, types.Code{},
			DefaultGasLimit, DefaultSendValue,
			[]types.TokenBalance{}, to, types.ExecutionTransactionKind))
	require.NoError(t, m.Sign(MainPrivateKey))

	return m
}

func Deploy(t *testing.T, ctx context.Context, es *ExecutionState,
	payload types.DeployPayload, shardId types.ShardId, from types.Address, seqno types.Seqno,
) types.Address {
	t.Helper()

	txn := NewDeployTransaction(payload, shardId, from, seqno, types.Value{})
	es.AddInTransaction(txn)
	execResult := es.HandleTransaction(ctx, txn, dummyPayer{})
	require.False(t, execResult.Failed())
	es.AddReceipt(execResult)

	return txn.To
}
