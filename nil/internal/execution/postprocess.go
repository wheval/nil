package execution

import (
	"errors"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
)

func PostprocessBlock(tx db.RwTx, shardId types.ShardId, blockResult *BlockGenerationResult) error {
	if blockResult.Block == nil {
		return errors.New("block is not set")
	}
	postprocessor := blockPostprocessor{tx, shardId, blockResult}
	return postprocessor.Postprocess()
}

type blockPostprocessor struct {
	tx          db.RwTx
	shardId     types.ShardId
	blockResult *BlockGenerationResult
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
	return db.WriteLastBlockHash(pp.tx, pp.shardId, pp.blockResult.BlockHash)
}

func (pp *blockPostprocessor) fillBlockHashByNumberIndex() error {
	return pp.tx.PutToShard(pp.shardId, db.BlockHashByNumberIndex, pp.blockResult.Block.Id.Bytes(), pp.blockResult.BlockHash.Bytes())
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
			blockHashAndTransactionIndex := db.BlockHashAndTransactionIndex{BlockHash: pp.blockResult.BlockHash, TransactionIndex: kv.Key}
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

	if err := fill(pp.blockResult.Block.InTransactionsRoot, db.BlockHashAndInTransactionIndexByTransactionHash); err != nil {
		return err
	}
	return fill(pp.blockResult.Block.OutTransactionsRoot, db.BlockHashAndOutTransactionIndexByTransactionHash)
}
