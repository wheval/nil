package ibft

import (
	"context"
	"sync"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/internal/types"
)

func (i *backendIBFT) calcProposer(height, round uint64) (*config.ValidatorInfo, error) {
	validators, err := i.validatorsCache.getValidators(i.ctx, height)
	if err != nil {
		i.logger.Error().
			Err(err).
			Uint64(logging.FieldRound, round).
			Uint64(logging.FieldHeight, height).
			Msg("Failed to get validators")
		return nil, err
	}

	index := (height + round) % uint64(len(validators))
	return &validators[index], nil
}

type validatorsMap struct {
	txFabtic db.DB

	shardId types.ShardId

	m sync.Map
}

func newValidatorsMap(txFabric db.DB, shardId types.ShardId) *validatorsMap {
	return &validatorsMap{
		txFabtic: txFabric,
		shardId:  shardId,
	}
}

func (m *validatorsMap) getValidators(ctx context.Context, height uint64) ([]config.ValidatorInfo, error) {
	vAny, _ := m.m.LoadOrStore(height, &validatorValue{
		txFabric: m.txFabtic,
		shardId:  m.shardId,
		height:   height,
	})
	v, _ := vAny.(*validatorValue)
	v.init(ctx)
	if v.err != nil {
		// This is likely to happen if we try to get validators for a height that is not yet available.
		// In this case, we should not cache the error, because the error is not permanent.
		m.m.Delete(height)
	}
	return v.value, v.err
}

type validatorValue struct {
	txFabric db.DB

	shardId types.ShardId
	height  uint64

	value []config.ValidatorInfo
	err   error

	once sync.Once
}

func (v *validatorValue) getValidators(ctx context.Context) ([]config.ValidatorInfo, error) {
	tx, err := v.txFabric.CreateRoTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var mainBlockHash *common.Hash

	if v.shardId.IsMainShard() {
		hash, err := db.ReadBlockHashByNumber(tx, v.shardId, types.BlockNumber(max(0, int64(v.height)-2)))
		if err != nil {
			return nil, err
		}
		mainBlockHash = &hash
	} else {
		block, err := db.ReadBlockByNumber(tx, v.shardId, types.BlockNumber(v.height-1))
		if err != nil {
			return nil, err
		}
		mainBlockHash = &block.MainChainHash
	}

	configAccessor, err := config.NewConfigAccessorTx(ctx, tx, mainBlockHash)
	if err != nil {
		return nil, err
	}

	validatorsList, err := config.GetParamValidators(configAccessor)
	if err != nil {
		return nil, err
	}

	return validatorsList.Validators[v.shardId].List, nil
}

func (v *validatorValue) init(ctx context.Context) {
	v.once.Do(func() {
		v.value, v.err = v.getValidators(ctx)
	})
}
