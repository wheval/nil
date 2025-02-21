package ibft

import (
	"context"
	"sync"

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
	return config.GetValidatorListForShard(ctx, v.txFabric, types.BlockNumber(v.height), v.shardId)
}

func (v *validatorValue) init(ctx context.Context) {
	v.once.Do(func() {
		v.value, v.err = v.getValidators(ctx)
	})
}
