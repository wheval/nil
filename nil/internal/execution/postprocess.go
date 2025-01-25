package execution

import (
	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
)

func PostprocessBlock(tx db.RwTx, shardId types.ShardId, defaultGasPrice types.Value, blockHash common.Hash) (*types.Block, error) {
	postprocessor, err := newBlockPostprocessor(tx, shardId, defaultGasPrice, blockHash)
	if err != nil {
		return nil, err
	}
	return postprocessor.block, postprocessor.Postprocess()
}

type blockPostprocessor struct {
	tx              db.RwTx
	shardId         types.ShardId
	blockHash       common.Hash
	block           *types.Block
	defaultGasPrice types.Value
}

func newBlockPostprocessor(tx db.RwTx, shardId types.ShardId, defaultGasPrice types.Value, blockHash common.Hash) (*blockPostprocessor, error) {
	block, err := db.ReadBlock(tx, shardId, blockHash)
	if err != nil {
		return nil, err
	}
	return &blockPostprocessor{tx, shardId, blockHash, block, defaultGasPrice}, nil
}

func (pp *blockPostprocessor) Postprocess() error {
	for _, postpocessor := range []func() error{
		pp.fillLastBlockTable,
		pp.fillBlockHashByNumberIndex,
		pp.fillBlockHashAndTransactionIndexByTransactionHash,
	} {
		if err := postpocessor(); err != nil {
			return err
		}
	}
	return nil
}

func (pp *blockPostprocessor) fillLastBlockTable() error {
	return db.WriteLastBlockHash(pp.tx, pp.shardId, pp.blockHash)
}

func (pp *blockPostprocessor) fillBlockHashByNumberIndex() error {
	if err := pp.tx.PutToShard(pp.shardId, db.BlockHashByNumberIndex, pp.block.Id.Bytes(), pp.blockHash.Bytes()); err != nil {
		return err
	}
	return nil
}

func (pp *blockPostprocessor) fillBlockHashAndTransactionIndexByTransactionHash() error {
	fill := func(root common.Hash, table db.ShardedTableName) error {
		mptTransactions := NewDbTransactionTrieReader(pp.tx, pp.shardId)
		mptTransactions.SetRootHash(root)
		txns, err := mptTransactions.Entries()
		if err != nil {
			return err
		}

		for _, kv := range txns {
			blockHashAndTransactionIndex := db.BlockHashAndTransactionIndex{BlockHash: pp.blockHash, TransactionIndex: kv.Key}
			value, err := blockHashAndTransactionIndex.MarshalSSZ()
			if err != nil {
				return err
			}

			if err := pp.tx.PutToShard(pp.shardId, table, kv.Val.Hash().Bytes(), value); err != nil {
				return err
			}
		}
		return nil
	}

	if err := fill(pp.block.InTransactionsRoot, db.BlockHashAndInTransactionIndexByTransactionHash); err != nil {
		return err
	}
	return fill(pp.block.OutTransactionsRoot, db.BlockHashAndOutTransactionIndexByTransactionHash)
}
